package engine

import (
	"bytes"
	"strings"

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
		log.Errorf("Command returned '%v': %s", err, out.String())
	}
	return err
}

func (cli *dockerCli) Build(dockerfile, image, dir string) error {
	return cli.cmd("build", "-f", dockerfile, "-t", image, dir)
}

func (cli *dockerCli) Name() string {
	return "docker"
}

func (cli *dockerCli) Push(image string) error {
	return cli.cmd("push", image)
}

func (cli *dockerCli) Pull(image string) error {
	return cli.cmd("pull", image)
}

func (cli *dockerCli) Tag(src, dst string) error {
	return cli.cmd("tag", src, dst)
}

func (cli *dockerCli) Version() (string, error) {
	var out bytes.Buffer
	err := cli.exec.NewCommand("docker", "version", "--format", "{{json .Server.Version}}").WithCombinedOutput(&out).Run()
	if err != nil {
		log.Errorf("Command returned '%v': %s", err, out.String())
	}
	return strings.ReplaceAll(strings.TrimSpace(out.String()), "\"", ""), nil
}
