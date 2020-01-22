package config

import (
	"fmt"
	"os"
	"path"
)

func NewBuilderDefinitionFilesystem(source, name string) (*BuilderDef, error) {
	fullPath := path.Join(source, name)
	if !pathExists(fullPath) {
		return nil, fmt.Errorf("A builder named '%s' could not be found at '%s'", name, source)
	}

	return &BuilderDef{
		source: path.Join(source, name),
	}, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
