package config

import (
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"

	log "github.com/sirupsen/logrus"
)

func NewBuilderDefinitionGit(source, name string) (*BuilderDef, error) {
	cacheRoot, err := getCacheRoot()
	if err != nil {
		return nil, err
	}

	sum := fmt.Sprintf("%x", md5.Sum([]byte(source)))

	cachePath := path.Join(cacheRoot, sum)
	log.Infof("Will clone repository into '%s'", cachePath)
	_, err = os.Stat(cachePath)
	if err == nil {
		cmd := exec.Command("git", "reset", "--hard", "origin/master")
		cmd.Dir = cachePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
	} else {
		cmd := exec.Command("git", "clone", source, cachePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
	}
	return &BuilderDef{
		source: path.Join(cacheRoot, sum, name),
	}, nil
}

func getCacheRoot() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	cacheRoot := fmt.Sprintf("%s/.image-builder/cache", usr.HomeDir)
	os.MkdirAll(cacheRoot, 0644)
	return cacheRoot, nil
}
