package engine

import "fmt"

type BuildEngine interface {
	Build(dockerfile, image, context string) error
	Push(image string) error
	Pull(image string) error
}

func New(name string) (BuildEngine, error) {
	if name == "podman" {
		return NewPodmanCli(), nil
	} else if name == "docker" {
		return NewDockerCli(), nil
	} else {
		return nil, fmt.Errorf("Unsupport engine")
	}
}
