package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
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

func simpleHealthCheck(c *cli.Context) {
	context := NewHealthContext(c)
	logrus.Infof("Listening on port: %s", context.port)

	http.Handle("/ping", context)
	err := http.ListenAndServe(fmt.Sprintf(":%s", context.port), nil)
	if err != nil {
		logrus.Fatal(err)
	}
}

type Response struct {
	Type   string `json:"type"`
	Status int    `json:"status"`
	Code   string `json:"code"`
}

func (h *HealthContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	message := "OK"
	code := 200

	if err := runCommand(h.checkCommand); err != nil {
		code = 503
		message = "Failed Health Check. Running: " + h.failureCommand
		if err = runCommand(h.failureCommand); err != nil {
			message += "....[Failed]"
		}
		message += "....[Success]"
	}
	response, _ := json.Marshal(getResponse(message, code))

	fmt.Fprintf(w, string(response))
}

func getResponse(msg string, code int) *Response {
	return &Response{
		Type:   msg,
		Status: code,
		Code:   http.StatusText(code),
	}
}

func runCommand(command string) error {
	if command != "" {
		cmd := exec.Command(command)
		return cmd.Run()
	}
	return nil
}
