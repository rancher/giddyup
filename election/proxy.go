package election

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
)

type Watcher struct {
	leader  metadata.Container
	command []string
	port    int
	dstPort int
	client  metadata.Client
	forward *TcpProxy
}

func New(client metadata.Client, port int, command []string) *Watcher {
	return &Watcher{
		command: command,
		port:    port,
		client:  client,
	}
}

func NewSrcDstWatcher(client metadata.Client, srcPort, dstPort int) *Watcher {
	return &Watcher{
		command: []string{},
		port:    srcPort,
		dstPort: dstPort,
		client:  client,
	}
}

func (w *Watcher) GetSelfServiceLeader() (metadata.Container, bool, error) {
	return w.getLeader()
}

func (w *Watcher) getLeader() (metadata.Container, bool, error) {
	selfContainer, err := w.client.GetSelfContainer()
	if err != nil {
		return metadata.Container{}, false, err
	}

	index := selfContainer.CreateIndex
	leader := selfContainer

	containers, err := w.client.GetServiceContainers(
		selfContainer.ServiceName,
		selfContainer.StackName,
	)
	if err != nil {
		return metadata.Container{}, false, err
	}

	for _, container := range containers {
		if container.CreateIndex < index {
			index = container.CreateIndex
			leader = container
		}
	}

	w.leader = leader
	return leader, leader.UUID == selfContainer.UUID, nil
}

func (w *Watcher) Forwarder() error {
	//initialize leader
	if _, _, err := w.getLeader(); err != nil {
		return err
	}

	w.forward = NewTcpProxy(w.port, func() string {
		return fmt.Sprintf("%s:%d", w.leader.PrimaryIp, w.dstPort)
	})

	go w.client.OnChange(1, func(version string) {
		currentLeaderIp := w.leader.PrimaryIp
		if _, _, err := w.getLeader(); err != nil {
			logrus.Error("Error getting leader: %s", err)
		}

		if w.leader.PrimaryIp != currentLeaderIp {
			if err := w.forward.Close(); err == nil {
				w.forward.Reset()
			}
		}
	})

	for {
		if w.port > 0 && w.dstPort > 0 {
			if err := w.forward.Forward(); err != nil {
				return err
			}
		} else {
			return errors.New("Ports not specified: src:" + string(w.port) + ", dst:" + string(w.dstPort))
		}
	}

	return errors.New("Unexpected loop termination")
}

func (w *Watcher) Watch() error {
	w.forward = NewTcpProxy(w.port, func() string {
		return fmt.Sprintf("%s:%d", w.leader.PrimaryIp, w.port)
	})

	go w.client.OnChange(2, func(version string) {
		if w.IsLeader() {
			w.forward.Close()
		}
	})

	if w.port > 0 {
		if err := w.forward.Forward(); err != nil {
			return err
		}
	}

	if w.IsLeader() {
		if len(w.command) == 0 {
			return errors.New("No command")
		}

		prog, err := exec.LookPath(w.command[0])
		if err != nil {
			return err
		}
		return syscall.Exec(prog, w.command, os.Environ())
	}

	return errors.New("Unexpected loop termination")
}

func (w *Watcher) IsLeader() bool {
	_, leader, err := w.getLeader()
	return leader && err == nil
}
