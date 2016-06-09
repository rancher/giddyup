package app

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cloudnautique/giddyup/election"
	"github.com/urfave/cli"
	"github.com/rancher/go-rancher-metadata/metadata"
)

var (
	port    = "proxy-tcp-port"
	dstPort = "dst-port"
	srcPort = "src-port"
	check   = "check"
)

func LeaderCommand() cli.Command {
	return cli.Command{
		Name:  "leader",
		Usage: "Provides a deterministic way to elect, route traffic, and get a leader of a service",
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
			{
				Name:   "forward",
				Usage:  "Listen and forward all port traffic to leader.",
				Action: appActionForward,
				Flags: []cli.Flag{
					cli.IntFlag{
						Name:  dstPort,
						Usage: "Leader destination port",
					},
					cli.IntFlag{
						Name:  srcPort,
						Usage: "Local source port",
					},
				},
			},
			{
				Name:   "get",
				Usage:  "Get the leader of service",
				Action: appActionGet,
			},
		},
	}
}

func appActionCheck(cli *cli.Context) error {
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
	return nil
}

func appActionGet(cli *cli.Context) error {
	client, err := metadata.NewClientAndWait(metadataURL)
	if err != nil {
		logrus.Fatal(err)
	}

	w := election.New(client, cli.Int(port), cli.Args())

	leader, _, err := w.GetSelfServiceLeader()
	if err != nil {
		logrus.Fatalf("Could not get leader. %s", err)
	}
	fmt.Printf("%s", leader.PrimaryIp)
	os.Exit(0)
	return nil
}

func appActionForward(cli *cli.Context) error {
	client, err := metadata.NewClientAndWait(metadataURL)
	if err != nil {
		logrus.Fatal(err)
	}

	dst := cli.Int(dstPort)

	if dst == 0 {
		dst = cli.Int(srcPort)
	}

	w := election.NewSrcDstWatcher(client, cli.Int(srcPort), dst)
	if err := w.Forwarder(); err != nil {
		logrus.Fatal(err)
	}
	return nil
}

func appActionElect(cli *cli.Context) error {
	client, err := metadata.NewClientAndWait(metadataURL)
	if err != nil {
		logrus.Fatal(err)
	}

	w := election.New(client, cli.Int(port), cli.Args())
	if err := w.Watch(); err != nil {
		logrus.Fatal(err)
	}
	return nil
}
