package app

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/go-rancher-metadata/metadata"
)

type timeoutError struct {
	message string
}

func (e *timeoutError) Error() string {
	return e.message
}

const metadataURL = "http://rancher-metadata/2015-07-25"

func ServiceCommand() cli.Command {
	return cli.Command{
		Name:  "service",
		Usage: "Service actions",
		Subcommands: []cli.Command{
			{
				Name:  "wait",
				Usage: "Wait for service container count to match set scale",
				Action: func(c *cli.Context) {
					if err := WaitForServiceScale(c.Int("timeout")); err != nil {
						logrus.Fatalf("Error: %v", err)
					}
				},
				Flags: []cli.Flag{
					cli.IntFlag{
						Name:  "timeout",
						Usage: "Time in seconds to wait for scale to be achieved. (Default: 600)",
						Value: 600,
					},
				},
			},
		},
	}
}

func WaitForServiceScale(timeout int) error {
	mdClient, err := metadata.NewClientAndWait(metadataURL)
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
