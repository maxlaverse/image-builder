package engine

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

type dockerCli struct{}

func NewDockerCli() BuildEngine {
	return &dockerCli{}
}

func (cli *dockerCli) cmd(args ...string) error {
	cmd := exec.Command("docker", args...)
	var out bytes.Buffer

	if log.GetLevel() >= log.InfoLevel {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &out
		cmd.Stderr = &out
	}
	log.Debugf("Executing docker %v", strings.Join(args, " "))
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
