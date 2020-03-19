package engine

import (
	"fmt"

	"github.com/maxlaverse/image-builder/pkg/executor"
)

// BuildEngine abstract container builder
type BuildEngine interface {
	Build(dockerfile, image, context string) error
	Name() string
	Push(image string) error
	Pull(image string) error
	Version() (string, error)
	Tag(src, dst string) error
}

// New returns a new container builder engine
func New(name string, exec executor.Executor) (BuildEngine, error) {
	if name == "podman" {
		return newPodmanCli(exec), nil
	} else if name == "docker" {
		return newDockerCli(exec), nil
	} else if name == "buildah" {
		return newbuildahCli(exec), nil
	} else {
		return nil, fmt.Errorf("Unsupport engine: %s", name)
	}
}
