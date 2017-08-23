package app

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/giddyup/election"
	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/urfave/cli"
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
				Usage:  "Get Rancher IP the leader of service. If you want, you can get can get underlying hostname or agent_ip",
				Action: appActionGet,
			},
		},
	}
}

func appActionCheck(cli *cli.Context) error {
	client, err := metadata.NewClientAndWait(cli.GlobalString("metadata-url"))
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
	client, err := metadata.NewClientAndWait(cli.GlobalString("metadata-url"))
	if err != nil {
		logrus.Fatal(err)
	}

	w := election.New(client, cli.Int(port), cli.Args())

	leader, _, err := w.GetSelfServiceLeader()
	if err != nil {
		logrus.Fatalf("Could not get leader. %s", err)
	}

	switch {
	case len(cli.Args()) == 0:
		fmt.Printf("%s", leader.PrimaryIp)
		os.Exit(0)

	case cli.Args()[0] == "host":
		host, err := client.GetHost(leader.HostUUID)
		if err != nil {
			return err
		}

		fmt.Printf("%s", host.Hostname)
		os.Exit(0)

	case cli.Args()[0] == "agent_ip":
		host, err := client.GetHost(leader.HostUUID)
		if err != nil {
			return err
		}

		fmt.Printf("%s", host.AgentIP)
		os.Exit(0)

	}

	return fmt.Errorf("Unrecognized arg: (%s) nothing, host and agent_ip are only allowed args", cli.Args()[0])
}

func appActionForward(cli *cli.Context) error {
	client, err := metadata.NewClientAndWait(cli.GlobalString("metadata-url"))
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
	client, err := metadata.NewClientAndWait(cli.GlobalString("metadata-url"))
	if err != nil {
		logrus.Fatal(err)
	}

	w := election.New(client, cli.Int(port), cli.Args())
	if err := w.Watch(); err != nil {
		logrus.Fatal(err)
	}
	return nil
}
