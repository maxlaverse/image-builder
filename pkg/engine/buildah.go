package engine

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/maxlaverse/image-builder/pkg/executor"
	log "github.com/sirupsen/logrus"
)

type buildahCli struct {
	exec executor.Executor
}

// newBuildahCli returns a new engine based on buildah
func newBuildahCli(exec executor.Executor) BuildEngine {
	return &buildahCli{exec: exec}
}

func (cli *buildahCli) cmd(args ...string) error {
	cmd := cli.exec.NewCommand("buildah", args...)
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

func (cli *buildahCli) Build(dockerfile, image, context string) error {
	return cli.cmd("build-using-dockerfile", "-f", dockerfile, "-t", image, context)
}

func (cli *buildahCli) Name() string {
	return "buildah"
}

func (cli *buildahCli) Push(image string) error {
	return cli.cmd("push", image)
}

func (cli *buildahCli) Pull(image string) error {
	return cli.cmd("pull", image)
}

func (cli *buildahCli) Tag(src, dst string) error {
	return cli.cmd("tag", src, dst)
}

func (cli *buildahCli) Version() (string, error) {
	var out bytes.Buffer

	err := cli.exec.NewCommand("buildah", "version").WithCombinedOutput(&out).Run()
	if err != nil {
		return "", fmt.Errorf("command returned '%v': %s", err, out.String())
	}
	return strings.TrimSpace(strings.Split(strings.Split(out.String(), "\n")[0], ":")[1]), nil
}
