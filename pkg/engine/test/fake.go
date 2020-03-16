package test

import (
	"fmt"
	"sync"
)

type fakeCli struct {
	MethodCalls   []string
	BuildCallback func(string)
	mux           sync.Mutex
}

// New returns a new engine based on Docker
func New() *fakeCli {
	return &fakeCli{}
}

// NewWithCallbacks returns a new engine based on Docker with callbacks on operations
func NewWithCallbacks(buildCb func(string)) *fakeCli {
	return &fakeCli{
		BuildCallback: buildCb,
	}
}

func (cli *fakeCli) Build(dockerfile, image, context string) error {
	if cli.BuildCallback != nil {
		cli.BuildCallback(image)
	}
	cli.mux.Lock()
	defer cli.mux.Unlock()
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("Build(%s)", image))
	return nil
}

func (cli *fakeCli) Push(image string) error {
	cli.mux.Lock()
	defer cli.mux.Unlock()
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("Push(%s)", image))
	return nil
}

func (cli *fakeCli) Pull(image string) error {
	cli.mux.Lock()
	defer cli.mux.Unlock()
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("Pull(%s)", image))
	return nil
}

func (cli *fakeCli) Tag(src, dst string) error {
	cli.mux.Lock()
	defer cli.mux.Unlock()
	cli.MethodCalls = append(cli.MethodCalls, fmt.Sprintf("Tag(%s,%s)", src, dst))
	return nil
}
