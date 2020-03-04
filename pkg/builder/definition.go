package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/maxlaverse/image-builder/pkg/builder/source"
	"github.com/maxlaverse/image-builder/pkg/utils"
)

// Definition is the interface to a builder definition, allowing to find stages
// or Dockerfiles to build
type Definition interface {
	CheckValidity() error
	GetStages() ([]string, error)
	GetStageDirectory(stage string) string
	GetStageDockerfile(stageName string) string
}

type builderDef struct {
	name string
	path string
}

// NewDefinitionFromLocation returns a builder definition from a local or
// remote location
func NewDefinitionFromLocation(name, location string) (Definition, error) {
	cacheRoot, err := getCacheRoot()
	if err != nil {
		return nil, err
	}

	var localPath string
	if source.IsSourceGit(location) {
		localPath, err = source.FromGit(name, location, cacheRoot)
	} else {
		localPath, err = source.FromFilesystem(name, location)
	}
	if err != nil {
		return nil, err
	}

	if !utils.PathExists(localPath) {
		return nil, fmt.Errorf("Builder '%s' was not found at '%s'", name, location)
	}

	def := NewDefinitionFromPath(name, localPath)
	if err := def.CheckValidity(); err != nil {
		return nil, err
	}

	return def, nil
}

// NewDefinitionFromPath returns a builder definition from the local path
func NewDefinitionFromPath(name, localPath string) Definition {
	return &builderDef{
		name: name,
		path: localPath,
	}
}

// CheckValidity returns if a builder seems valid
func (b *builderDef) CheckValidity() error {
	stages, err := b.GetStages()
	if err != nil {
		return err
	}

	if len(stages) == 0 {
		return fmt.Errorf("No stages found for Builder '%s'", b.name)
	}
	return nil
}

// GetStages returns the name of the supported stages
func (b *builderDef) GetStages() ([]string, error) {
	files, err := ioutil.ReadDir(b.path)
	if err != nil {
		return nil, err
	}

	stages := []string{}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		_, err = os.Stat(path.Join(b.path, file.Name(), "Dockerfile"))
		if err != nil {
			return nil, err
		}
		stages = append(stages, file.Name())
	}
	return stages, nil
}

// GetStageDirectory returns the path of the folder of a stage
func (b *builderDef) GetStageDirectory(stageName string) string {
	return path.Join(b.path, stageName)
}

// GetStageDockerfile returns the path of the folder of a stage
func (b *builderDef) GetStageDockerfile(stageName string) string {
	return path.Join(b.path, stageName, "Dockerfile")
}

func getCacheRoot() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	// TODO: Make this configurable
	cacheRoot := fmt.Sprintf("%s/.image-builder/cache", usr.HomeDir)
	os.MkdirAll(cacheRoot, 0644)
	return cacheRoot, nil
}
