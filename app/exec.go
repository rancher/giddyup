package app

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"

	"os/exec"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

func ExecCommand() cli.Command {
	return cli.Command{
		Name:   "exec",
		Usage:  "exec out to a command",
		Action: execCommand,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "secret-envs",
				Usage: "reads /run/secrets and sets env vars",
			},
		},
	}
}

func execCommand(c *cli.Context) error {
	if c.Bool("secret-envs") {
		envs, err := filesToMap("/run/secrets")
		if err != nil {
			logrus.Error(err)
			return err
		}

		for key, val := range envs {
			os.Setenv(strings.ToUpper(key), val)
		}
	}

	name, err := exec.LookPath(c.Args().Get(0))
	if err != nil {
		return err
	}

	return syscall.Exec(name, c.Args(), os.Environ())
}

func filesToMap(dirPath string) (map[string]string, error) {
	vals := map[string]string{}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return vals, err
	}

	for _, file := range files {
		content, err := ioutil.ReadFile(path.Join(dirPath, file.Name()))
		if err != nil {
			return vals, err
		}
		vals[file.Name()] = string(content)
	}
	return vals, nil
}
