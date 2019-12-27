package main

import (
	"os"

	giddyupApp "github.com/johnnyisme/giddyup/app"
	"github.com/johnnyisme/giddyup/version"
	"github.com/urfave/cli"
)

const metadataURL = "http://rancher-metadata/2015-12-19"

func main() {
	app := cli.NewApp()
	app.Version = version.VERSION
	app.Name = "giddyup"
	app.Usage = "Entrypoint functions for Rancher"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "metadata-url",
			Usage: "override default MetadataURL",
			Value: metadataURL,
		},
	}

	app.Commands = []cli.Command{
		giddyupApp.ExecCommand(),
		giddyupApp.HealthCommand(),
		giddyupApp.IPCommand(),
		giddyupApp.LeaderCommand(),
		giddyupApp.ProbeCommand(),
		giddyupApp.ServiceCommand(),
	}

	app.Run(os.Args)
}
