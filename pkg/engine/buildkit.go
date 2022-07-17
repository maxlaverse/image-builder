package engine

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/maxlaverse/image-builder/pkg/executor"
	log "github.com/sirupsen/logrus"
)

const (
	buildctlBinary = "buildctl"
)

type buildkitCli struct {
	exec executor.Executor
}

// newBuildkitCli returns a new engine based on Buildkit
func newBuildkitCli(exec executor.Executor) BuildEngine {
	return &buildkitCli{exec: exec}
}

func (cli *buildkitCli) cmd(args ...string) error {
	cmd := cli.exec.NewCommand(buildctlBinary, args...)
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

func (cli *buildkitCli) Build(buildkitfile, image, dir string) error {
	return cli.BuildAndPush(buildkitfile, image, dir, false)
}

func (cli *buildkitCli) BuildAndPush(buildkitfile, image, dir string, push bool) error {
	var pushStr string
	if push {
		pushStr = "true"
	} else {
		pushStr = "false"
	}
	return cli.cmd("build", "--frontend=dockerfile.v0", "--local", fmt.Sprintf("dockerfile=%s", filepath.Dir(buildkitfile)), "--opt", fmt.Sprintf("filename=%s", filepath.Base(buildkitfile)), "--local", fmt.Sprintf("context=%s", dir), "--output", fmt.Sprintf("type=image,name=%s,push=%s", image, pushStr))
}

func (cli *buildkitCli) Name() string {
	return "buildkit"
}

func (cli *buildkitCli) Push(image string) error {
	return fmt.Errorf("not supported")
}

func (cli *buildkitCli) Pull(image string) error {
	return fmt.Errorf("not supported")
}

func (cli *buildkitCli) Tag(src, dst string) error {
	return cli.cmd("tag", src, dst)
}

func (cli *buildkitCli) Version() (string, error) {
	var out bytes.Buffer
	err := cli.exec.NewCommand(buildctlBinary, "--version").WithCombinedOutput(&out).Run()
	if err != nil {
		return "", fmt.Errorf("command returned '%v': %s", err, out.String())
	}
	return strings.TrimSuffix(strings.Join(strings.Split(out.String(), " ")[1:], " "), "\n"), nil
}
