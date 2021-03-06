package engine

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/maxlaverse/image-builder/pkg/executor"
	log "github.com/sirupsen/logrus"
)

type podmanCli struct {
	exec executor.Executor
}

// newPodmanCli returns a new engine based on Podman
func newPodmanCli(exec executor.Executor) BuildEngine {
	return &podmanCli{exec: exec}
}

func (cli *podmanCli) cmd(args ...string) error {
	cmd := cli.exec.NewCommand("podman", args...)
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

func (cli *podmanCli) Build(dockerfile, image, context string) error {
	return cli.cmd("build", "--format=docker", "--cgroup-manager", "cgroupfs", "-f", dockerfile, "-t", image, context)
}

func (cli *podmanCli) Name() string {
	return "podman"
}

func (cli *podmanCli) Push(image string) error {
	return cli.cmd("push", image)
}

func (cli *podmanCli) Pull(image string) error {
	return cli.cmd("pull", image)
}

func (cli *podmanCli) Tag(src, dst string) error {
	return cli.cmd("tag", src, dst)
}

func (cli *podmanCli) Version() (string, error) {
	var out bytes.Buffer

	err := cli.exec.NewCommand("podman", "version", "--format", "{{.Server.Version}}").WithCombinedOutput(&out).Run()
	if err != nil {
		return "", fmt.Errorf("command returned '%v': %s", err, out.String())
	}
	return strings.ReplaceAll(strings.TrimSpace(out.String()), "\"", ""), nil
}
