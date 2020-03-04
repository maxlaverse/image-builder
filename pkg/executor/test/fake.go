package test

import (
	"fmt"

	"github.com/maxlaverse/image-builder/pkg/executor"
)

type fakeExecutor struct {
	MethodCalls []string
}

// New returns a new engine based on Docker
func New() *fakeExecutor {
	return &fakeExecutor{}
}

func (cli *fakeExecutor) NewCommand(cmd string, args ...string) executor.Command {
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("NewCommand(%s,%s)", cmd, args))
	return nil
}

func (cli *fakeExecutor) Build(dockerfile, image, context string) error {
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("Build(%s)", image))
	return nil
}

func (cli *fakeExecutor) Push(image string) error {
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("Push(%s)", image))
	return nil
}

func (cli *fakeExecutor) Pull(image string) error {
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("Pull(%s)", image))
	return nil
}

func (cli *fakeExecutor) Tag(src, dst string) error {
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("Tag(%s,%s)", src, dst))
	return nil
}
