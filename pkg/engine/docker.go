package engine

import (
	"bytes"

	"github.com/maxlaverse/image-builder/pkg/executor"
	log "github.com/sirupsen/logrus"
)

type dockerCli struct {
	exec executor.Executor
}

// newDockerCli returns a new engine based on Docker
func newDockerCli(exec executor.Executor) BuildEngine {
	return &dockerCli{exec: exec}
}

func (cli *dockerCli) cmd(args ...string) error {
	cmd := cli.exec.NewCommand("docker", args...)
	var out bytes.Buffer

	if log.GetLevel() >= log.InfoLevel {
		cmd = cmd.WithConsoleOutput()
	} else {
		cmd = cmd.WithCombinedOutput(&out)
	}
	err := cmd.Run()
	if err != nil {
		log.Errorf("Command returned: %s", out.String())
	}
	return err
}

func (cli *dockerCli) Build(dockerfile, image, context string) error {
	return cli.cmd("build", "-f", dockerfile, "-t", image, context)
}

func (cli *dockerCli) Push(image string) error {
	return cli.cmd("push", image)
}

func (cli *dockerCli) Pull(image string) error {
	return cli.cmd("pull", image)
}
