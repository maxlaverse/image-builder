package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// BuilderDef holds the struct for a builder
type BuilderDef struct {
	path string
}

// NewBuilderDef returns a new BuilderDef instance
func NewBuilderDef(location, name string) (*BuilderDef, error) {
	var err error
	var buildDef *BuilderDef

	if IsSourceGit(location) {
		buildDef, err = NewBuilderDefinitionGit(location, name)
	} else {
		buildDef, err = NewBuilderDefinitionFilesystem(location, name)
	}

	if err != nil {
		return nil, err
	}
	if len(buildDef.GetStages()) == 0 {
		return nil, fmt.Errorf("No stages found in '%s'", buildDef.path)
	}
	return buildDef, err
}

// GetStages returns the name of the stages available for the builder
func (b *BuilderDef) GetStages() []string {
	files, err := ioutil.ReadDir(b.path + "/")
	if err != nil {
		panic(err)
	}

	list := []string{}
	for _, file := range files {
		if file.IsDir() {
			_, err = os.Stat(path.Join(b.path, file.Name(), "Dockerfile"))
			if err == nil {
				list = append(list, file.Name())
			}
		}
	}
	return list
}

// GetStagePath returns the path of the folder of a stage
func (b *BuilderDef) GetStagePath(stage string) string {
	return path.Join(b.path, stage)
}
