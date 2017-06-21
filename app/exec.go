package app

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

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
			cli.StringSliceFlag{
				Name:  "wait-for-file",
				Usage: "wait for a file to exist, assumes something else is creating it. This flag can be used more then once for multiple files",
			},
			cli.StringSliceFlag{
				Name:  "source-file",
				Usage: "Source an environment file before executing. Can use the flag multiple times",
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

	if len(c.StringSlice("wait-for-file")) > 0 {
		err := waitForFiles(c.StringSlice("wait-for-file"))
		if err != nil {
			return err
		}
	}

	if len(c.StringSlice("source-file")) > 0 {
		envs, err := readSourceFiles(c.StringSlice("source-file"))
		if err != nil {
			return err
		}

		for key, val := range envs {
			os.Setenv(key, val)
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

func waitForFiles(files []string) error {
	for true {
		seenCount := 0
		for idx, file := range files {
			if file != "seen" {
				if _, err := os.Stat(file); os.IsNotExist(err) {
					break
				}
				// set the seen value and increment the counter
				files[idx] = "seen"
				seenCount++

				continue
			}
			seenCount++
		}
		if seenCount == len(files) {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func readSourceFiles(files []string) (map[string]string, error) {
	envs := map[string]string{}
	for _, file := range files {
		// the Close() will not be deferred to avoid large number of open files...
		f, err := os.Open(file)
		if err != nil {
			return envs, err
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			pair := strings.SplitN(scanner.Text(), "=", 2)
			if len(pair) == 2 {
				envs[pair[0]] = pair[1]
			}
		}
		f.Close()
	}
	return envs, nil
}
