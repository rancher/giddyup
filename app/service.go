package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/urfave/cli"
)

// const metadataURL = "http://rancher-metadata/2015-12-19"

type timeoutError struct {
	message string
}

func (e *timeoutError) Error() string {
	return e.message
}

type containerCollection struct {
	containers []metadata.Container
}

func (c *containerCollection) removeEntry(entry string) {
	newSlice := []metadata.Container{}
	for _, item := range c.containers {
		if item.Name != entry {
			newSlice = append(newSlice, item)
		}
	}
	c.containers = newSlice
}

func (c *containerCollection) printContainers(delim string) {
	names := []string{}

	for _, container := range c.containers {
		names = append(names, container.Name)
	}

	fmt.Print(strings.Join(names, delim))
}

func ServiceCommand() cli.Command {
	return cli.Command{
		Name:  "service",
		Usage: "Service actions",
		Subcommands: []cli.Command{
			{
				Name:  "wait",
				Usage: "Wait for service states",
				Subcommands: []cli.Command{
					{
						Name:  "scale",
						Usage: "Wait for number of service containers to reach set scale",
						//Action: func(c *cli.Context) {
						//if err := WaitForServiceScale(c.Int("timeout")); err != nil {
						//logrus.Fatalf("Error: %v", err)
						//}
						//},
						Action: WaitForServiceScale,
						Flags: []cli.Flag{
							cli.IntFlag{
								Name:  "timeout",
								Usage: "Time in seconds to wait for scale to be achieved. (Default: 600)",
								Value: 600,
							},
						},
					},
				},
			},
			{
				Name:   "scale",
				Usage:  "Print the desired scale of the service (name may be specified as a positional argument)",
				Action: appActionGetScale,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "current",
						Usage: "Print the current scale of the service",
					},
				},
			},
			{
				Name:   "containers",
				Usage:  "lists containers in the calling container's service one per line",
				Action: appActionGetServiceContainers,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "n",
						Usage: "print space separated list",
					},
					cli.BoolFlag{
						Name:  "exclude-self",
						Usage: "do not include calling container name in returned list",
					},
				},
			},
		},
	}
}

func WaitForServiceScale(c *cli.Context) error {
	timeout := c.Int("timeout")
	mdClient, err := metadata.NewClientAndWait(c.GlobalString("metadata-url"))
	if err != nil {
		return err
	}

	service, err := mdClient.GetSelfService()
	if err != nil {
		return err
	}

	desiredScale := service.Scale
	currentScale := len(service.Containers)

	timeInc := 0
	for timeInc < timeout {
		service, err = mdClient.GetSelfService()
		if err != nil {
			return err
		}

		if currentScale = len(service.Containers); currentScale >= desiredScale {
			os.Exit(0)
		}

		time.Sleep(1 * time.Second)
		timeInc++
	}

	if timeInc >= timeout {
		return &timeoutError{"Timed out waiting for service: " + service.Name + " to reach scale"}
	}

	return nil
}

func appActionGetScale(c *cli.Context) error {
	client, err := metadata.NewClientAndWait(c.GlobalString("metadata-url"))
	if err != nil {
		logrus.Fatal(err)
	}

	var service metadata.Service

	name := c.Args().First()
	if name != "" {
		service, err = client.GetSelfServiceByName(name)
	} else {
		service, err = client.GetSelfService()
	}

	if err != nil {
		logrus.Fatal(err)
	}

	if c.Bool("current") {
		fmt.Printf("%d", len(service.Containers))
		os.Exit(0)
	}

	fmt.Print(service.Scale)
	return nil
}

func appActionGetServiceContainers(c *cli.Context) error {
	delimiter := "\n"
	client, err := metadata.NewClientAndWait(c.GlobalString("metadata-url"))
	if err != nil {
		logrus.Fatal(err)
	}

	service, _ := client.GetSelfService()

	if c.Bool("n") {
		delimiter = " "
	}

	containerCollection := &containerCollection{
		containers: service.Containers,
	}
	if c.Bool("exclude-self") {
		selfContainer, _ := client.GetSelfContainer()
		containerCollection.removeEntry(selfContainer.Name)
	}

	containerCollection.printContainers(delimiter)
	return nil
}
