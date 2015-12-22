package election

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

type TcpProxy struct {
	done chan bool
	port int
	to   func() string
}

func NewTcpProxy(port int, to func() string) *TcpProxy {
	return &TcpProxy{
		port: port,
		to:   to,
		done: make(chan bool),
	}
}

func (t *TcpProxy) Close() error {
	t.done <- true
	return nil
}

func (t *TcpProxy) Reset() error {
	t.done <- false
	return nil
}

func (t *TcpProxy) Forward() error {
	strAddr := fmt.Sprintf("0.0.0.0:%d", t.port)

	addr, err := net.ResolveTCPAddr("tcp", strAddr)
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	logrus.Infof("Listening on %s", strAddr)
	logrus.Infof("Forwarding setup to: %s", t.to())

	wg := sync.WaitGroup{}

	for {
		select {
		case <-t.done:
			return nil
		default:
			break
		}

		l.SetDeadline(time.Now().Add(1 * time.Second))
		conn, err := l.AcceptTCP()
		if acceptErr, ok := err.(*net.OpError); ok && acceptErr.Timeout() {
			continue
		}
		if err != nil {
			return err
		}

		wg.Add(1)
		go func(conn *net.TCPConn) {
			defer wg.Done()
			if err := t.forward(conn); err != nil {
				logrus.Errorf("Failed handling TCP forwarding: %v", err)
			}
		}(conn)
	}

	wg.Wait()
	return nil
}

func (t *TcpProxy) forward(cConn *net.TCPConn) error {
	//cConn is the incoming client connection.
	defer cConn.Close()
	ip := t.to()
	if ip == "" {
		return errors.New("Target unknown")
	}

	raddr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		return err
	}
	rConn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return err
	}
	defer rConn.Close()

	cClose := make(chan bool)
	rClose := make(chan bool)

	go connectionBroker(rConn, cConn, cClose)
	go connectionBroker(cConn, rConn, rClose)

	done := make(chan bool)
	select {
	case <-cClose:
		rConn.SetLinger(0)
		rConn.CloseRead()
		done = rClose
	case <-rClose:
		cConn.CloseRead()
		done = cClose
	}

	<-done

	return err
}

func connectionBroker(dst, src net.Conn, srcClosed chan bool) {
	_, err := io.Copy(dst, src)
	if err != nil {
		logrus.Error(err)
	}

	if err := src.Close(); err != nil {
		logrus.Errorf("Connection close error: %s", err)
	}

	srcClosed <- true
}
