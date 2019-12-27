package app

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func HealthCommand() cli.Command {
	return cli.Command{
		Name:   "health",
		Usage:  "simple healthcheck",
		Action: simpleHealthCheck,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "listen-port,p",
				Usage: "set port to listen on",
				Value: 1620,
			},
			cli.StringFlag{
				Name:  "check-command",
				Usage: "command to execute check",
			},
			cli.StringFlag{
				Name:  "on-failure-command",
				Usage: "command to execute if command fails",
			},
		},
	}
}

type HealthContext struct {
	port           string
	checkCommand   string
	failureCommand string
}

func NewHealthContext(c *cli.Context) *HealthContext {
	context := &HealthContext{}
	context.port = c.String("listen-port")
	context.checkCommand = c.String("check-command")
	context.failureCommand = c.String("on-failure-command")

	return context
}

func simpleHealthCheck(c *cli.Context) error {
	context := NewHealthContext(c)
	logrus.Infof("Listening on port: %s", context.port)

	http.Handle("/ping", context)
	done := make(chan error)

	go func() {
		done <- http.ListenAndServe(fmt.Sprintf(":%s", context.port), nil)
	}()

	if len(c.Args()) > 0 {
		go func() {
			done <- runCommand(c.Args()[0], c.Args()[1:]...)
		}()
	}

	err := <-done
	if err != nil {
		logrus.Fatal(err)
	}
	return nil
}

func (h *HealthContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	message := "OK"
	code := http.StatusOK

	if err := runCommand(h.checkCommand); err != nil {
		code = http.StatusServiceUnavailable
		message = "Failed Health Check. Attempting to Run: " + h.failureCommand
		cmdStatus := "....[Success]"
		if err = runCommand(h.failureCommand); err != nil {
			cmdStatus = "....[Failed]"
		}
		message += cmdStatus
	}

	w.WriteHeader(code)
	fmt.Fprintln(w, message)
}

func runCommand(command string, args ...string) error {
	if command != "" {
		cmd := exec.Command(command, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}
