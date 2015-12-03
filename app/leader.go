package app

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/go-rancher-metadata/metadata"
)

func LeaderCommand() cli.Command {
	return cli.Command{
		Name:  "leader",
		Usage: "Determines if this container has lowest start index",
		Action: func(c *cli.Context) {
			if err := LowestContainerCreateIndex(); err != nil {
				logrus.Fatalf("Error: %v", err)
			}
		},
	}
}

func LowestContainerCreateIndex() error {
	mdClient, err := metadata.NewClientAndWait(metadataURL)
	if err != nil {
		return err
	}

	selfContainer, err := mdClient.GetSelfContainer()
	if err != nil {
		return err
	}

	serviceContainers, err := mdClient.GetServiceContainers(
		selfContainer.ServiceName,
		selfContainer.StackName,
	)
	if err != nil {
		return err
	}

	for _, container := range serviceContainers {
		if selfContainer.CreateIndex > container.CreateIndex {
			os.Exit(1)
		}
	}
	os.Exit(0)

	return nil
}
