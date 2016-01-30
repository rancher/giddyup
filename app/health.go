package app

import (
	"fmt"
	"net/http"

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

type HealthHandler struct {
	ctx *HealthContext
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

func (h *HealthContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}
