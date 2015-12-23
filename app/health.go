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
		},
	}
}

func simpleHealthCheck(c *cli.Context) {
	port := c.String("listen-port")
	logrus.Infof("Listening on port: %s", port)

	http.HandleFunc("/ping", handler)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		logrus.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}
