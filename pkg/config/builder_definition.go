package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// BuilderDef holds the struct for a builder
type BuilderDef struct {
	source string
}

// NewBuilderDef returns a new BuilderDef instance
func NewBuilderDef(source, name string) (*BuilderDef, error) {
	var err error
	var buildDef *BuilderDef

	if strings.Contains(source, "http") || strings.Contains(source, "git@") {
		buildDef, err = NewBuilderDefinitionGit(source, name)
	} else {
		buildDef, err = NewBuilderDefinitionFilesystem(source, name)
	}

	if err != nil {
		return nil, err
	}
	if len(buildDef.GetStages()) == 0 {
		return nil, fmt.Errorf("No stages found in '%s'", buildDef.source)
	}
	return buildDef, err
}

// GetStages returns the name of the stages available for the builder
func (b *BuilderDef) GetStages() []string {
	files, err := ioutil.ReadDir(b.source + "/")
	if err != nil {
		panic(err)
	}

	list := []string{}
	for _, file := range files {
		if file.IsDir() {
			_, err = os.Stat(path.Join(b.source, file.Name(), "Dockerfile"))
			if err == nil {
				list = append(list, file.Name())
			}
		}
	}
	return list
}

// GetStagePath returns the path of the folder of a stage
func (b *BuilderDef) GetStagePath(stage string) string {
	return path.Join(b.source, stage)
}
