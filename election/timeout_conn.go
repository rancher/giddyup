package election

import (
	"net"
	"time"
)

type TimeoutConn struct {
	Conn    net.Conn
	Timeout time.Duration
}

func (i TimeoutConn) Read(buf []byte) (int, error) {
	i.Conn.SetDeadline(time.Now().Add(i.Timeout * time.Second))
	return i.Conn.Read(buf)
}

func (i TimeoutConn) Write(buf []byte) (int, error) {
	i.Conn.SetDeadline(time.Now().Add(i.Timeout * time.Second))
	return i.Conn.Write(buf)
}
