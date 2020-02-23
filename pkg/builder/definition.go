package builder

import (
	"io/ioutil"
	"os"
	"path"
)

// Definition holds the struct for a Builder Definition
type Definition interface {
	GetStages() []string
	GetStagePath(stage string) string
}

type builderDef struct {
	path string
}

// NewDefinitionFromPath returns a new definition
func NewDefinitionFromPath(path string) Definition {
	return &builderDef{
		path: path,
	}
}

// GetStages returns the name of the supported stages
func (b *builderDef) GetStages() []string {
	files, err := ioutil.ReadDir(path.Join(b.path, "/"))
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
func (b *builderDef) GetStagePath(stage string) string {
	return path.Join(b.path, stage)
}
