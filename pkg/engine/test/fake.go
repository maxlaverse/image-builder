package test

import "fmt"

type fakeCli struct {
	MethodCalls []string
}

// New returns a new engine based on Docker
func New() *fakeCli {
	return &fakeCli{}
}

func (cli *fakeCli) Build(dockerfile, image, context string) error {
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("Build(%s)", image))
	return nil
}

func (cli *fakeCli) Push(image string) error {
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("Push(%s)", image))
	return nil
}

func (cli *fakeCli) Pull(image string) error {
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("Pull(%s)", image))
	return nil
}

func (cli *fakeCli) Tag(src, dst string) error {
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("Tag(%s,%s)", src, dst))
	return nil
}
