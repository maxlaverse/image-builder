package source

import (
	"crypto/md5"
	"fmt"
	"path"
	"strings"

	"github.com/maxlaverse/image-builder/pkg/executor"
	"github.com/maxlaverse/image-builder/pkg/utils"
)

const (
	defaultBranch = "master"
)

// IsSourceGit returns wether a location is a Git repository or not
func IsSourceGit(location string) bool {
	return strings.Contains(location, "http") || strings.Contains(location, "git@")
}

// FromGit cache a builder definition hosted in a Git repository
// Format is ssh://git@github.com:maxlaverse/image-builder-collection.git[#branch:[subfolder]]
func FromGit(name, location, cacheRoot string) (string, error) {
	locationParts := strings.Split(location, "#")
	repository := locationParts[0]
	branch := defaultBranch
	subdirectory := ""
	if len(locationParts) > 1 {
		otherParts := strings.Split(locationParts[1], ":")
		branch = otherParts[0]
		if len(otherParts) > 1 {
			subdirectory = otherParts[1]
		}
	}

	cachePath := path.Join(cacheRoot, locationFingerprint(location))
	if utils.PathExists(cachePath) {
		err := executor.New().NewCommand("git", "fetch", "--all").WithDir(cachePath).WithConsoleOutput().Run()
		if err != nil {
			return "", err
		}
	} else {
		err := executor.New().NewCommand("git", "clone", repository, cachePath).WithConsoleOutput().Run()
		if err != nil {
			return "", err
		}
	}

	err := executor.New().NewCommand("git", "reset", "--hard", fmt.Sprintf("origin/%s", branch)).WithDir(cachePath).WithConsoleOutput().Run()
	if err != nil {
		return "", err
	}

	return path.Join(cachePath, subdirectory, name), nil
}

func locationFingerprint(location string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(location)))
}
