package main

import (
	"os"

	giddyupApp "github.com/rancher/giddyup/app"
	"github.com/rancher/giddyup/version"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = version.VERSION
	app.Name = "giddyup"
	app.Usage = "Entrypoint functions for Rancher"

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
