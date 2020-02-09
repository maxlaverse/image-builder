package config

import (
	"fmt"
	"os"
	"path"
)

func NewBuilderDefinitionFilesystem(location, name string) (*BuilderDef, error) {
	fullPath := path.Join(location, name)
	if !pathExists(fullPath) {
		return nil, fmt.Errorf("A builder named '%s' could not be found at '%s'", name, location)
	}

	return &BuilderDef{
		path: path.Join(location, name),
	}, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
