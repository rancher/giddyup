package main

import (
	"os"

	giddyupApp "github.com/cloudnautique/giddyup/app"
	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "giddyup"
	app.Usage = "Entrypoint functions for Rancher"

	app.Commands = []cli.Command{
		giddyupApp.ServiceCommand(),
		giddyupApp.LeaderCommand(),
	}

	app.Run(os.Args)
}
