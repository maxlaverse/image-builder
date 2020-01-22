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

func (cli *dockerCli) cmd(ignoreError bool, args ...string) error {
	cmd := exec.Command("docker", args...)
	var out bytes.Buffer
	if log.GetLevel() >= log.InfoLevel && !ignoreError {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &out
		cmd.Stderr = &out
	}
	err := cmd.Run()
	log.Debugf("Executing docker %v", strings.Join(args, " "))
	if err != nil && !ignoreError {
		log.Errorf(out.String())
	}
	return err
}

func (cli *dockerCli) Build(dockerfile, image, context string) error {
	return cli.cmd(false, "build", "-f", dockerfile, "-t", image, context)
}

func (cli *dockerCli) Push(image string) error {
	return cli.cmd(false, "push", image)
}

func (cli *dockerCli) Pull(image string) error {
	return cli.cmd(false, "pull", image)
}

func (cli *dockerCli) Exists(image string) bool {
	err := cli.cmd(true, "inspect", image)
	if err == nil {
		return true
	}
	err = cli.cmd(true, "pull", image)
	return err == nil
}
