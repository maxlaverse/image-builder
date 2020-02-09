package config

import (
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

func IsSourceGit(location string) bool {
	return strings.Contains(location, "http") || strings.Contains(location, "git@")
}

func NewBuilderDefinitionGit(location, name string) (*BuilderDef, error) {
	cacheRoot, err := getCacheRoot()
	if err != nil {
		return nil, err
	}

	parts := strings.Split(location, "#")
	repository := parts[0]
	branch := "master"
	subDirectory := ""
	if len(parts) > 1 {
		otherParts := strings.Split(parts[1], ":")
		branch = otherParts[0]
		if len(otherParts) > 1 {
			subDirectory = otherParts[1]
		}
	}
	sum := fmt.Sprintf("%x", md5.Sum([]byte(location)))

	cachePath := path.Join(cacheRoot, sum)
	log.Infof("Will clone repository into '%s'", cachePath)
	_, err = os.Stat(cachePath)
	if err == nil {
		cmd := exec.Command("git", "fetch", "--all")
		cmd.Dir = cachePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
	} else {
		cmd := exec.Command("git", "clone", repository, cachePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
	}

	cmd := exec.Command("git", "reset", "--hard", fmt.Sprintf("origin/%s", branch))
	cmd.Dir = cachePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return &BuilderDef{
		path: path.Join(cacheRoot, sum, subDirectory, name),
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
