package main

import (
	"os"

	giddyupApp "github.com/cloudnautique/giddyup/app"
	"github.com/cloudnautique/giddyup/version"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = version.VERSION
	app.Name = "giddyup"
	app.Usage = "Entrypoint functions for Rancher"

	app.Commands = []cli.Command{
		giddyupApp.IPCommand(),
		giddyupApp.LeaderCommand(),
		giddyupApp.ServiceCommand(),
		giddyupApp.HealthCommand(),
		giddyupApp.HealthzCommand(),
	}

	app.Run(os.Args)
}
