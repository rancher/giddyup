package app

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/rancher/leader/election"
)

var (
	port  = "proxy-tcp-port"
	check = "check"
)

func LeaderCommand() cli.Command {
	return cli.Command{
		Name:  "leader",
		Usage: "Determines if this container has lowest start index",
		Subcommands: []cli.Command{
			{
				Name:   "check",
				Usage:  "Check if we are leader and exit.",
				Action: appActionCheck,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "service",
						Usage: "Get the leader of another service in the stack",
					},
				},
			},
			{
				Name:   "elect",
				Usage:  "Simple leader election with Rancher",
				Action: appActionElect,
				Flags: []cli.Flag{
					cli.IntFlag{
						Name:  port,
						Usage: "Port to proxy to the leader",
					},
				},
			},
		},
	}
}

func appActionCheck(cli *cli.Context) {
	client, err := metadata.NewClientAndWait(metadataURL)
	if err != nil {
		logrus.Fatal(err)
	}

	w := election.New(client, cli.Int(port), cli.Args())

	if w.IsLeader() {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func appActionElect(cli *cli.Context) {
	client, err := metadata.NewClientAndWait(metadataURL)
	if err != nil {
		logrus.Fatal(err)
	}

	w := election.New(client, cli.Int(port), cli.Args())
	if err := w.Watch(); err != nil {
		logrus.Fatal(err)
	}
}
