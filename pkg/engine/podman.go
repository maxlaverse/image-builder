package engine

import (
	"io/ioutil"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type podmanCli struct {
}

func NewPodmanCli() BuildEngine {
	return &podmanCli{}
}

func (cli *podmanCli) Build(dockerfile, image, context string) error {
	cmd := exec.Command("podman", "build", "--format=docker", "--cgroup-manager", "cgroupfs", "-f", dockerfile, "-t", image, context)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		dat, _ := ioutil.ReadFile(dockerfile)
		log.Error(string(dat))
		return err
	}
	return nil
}

func (cli *podmanCli) Push(image string) error {
	cmd := exec.Command("podman", "push", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (cli *podmanCli) Pull(image string) error {
	cmd := exec.Command("podman", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (cli *podmanCli) Exists(image string) bool {
	cmd := exec.Command("podman", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}
