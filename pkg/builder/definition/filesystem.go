package definition

import (
	"fmt"
	"os"
	"path"

	"github.com/maxlaverse/image-builder/pkg/builder"
)

// FromFileSystem creates a new Builder Definition from a location on the filesystem
func FromFileSystem(location, name string) (builder.Definition, error) {
	fullPath := path.Join(location, name)
	if !pathExists(fullPath) {
		return nil, fmt.Errorf("A builder named '%s' could not be found at '%s'", name, location)
	}

	return builder.NewDefinitionFromPath(path.Join(location, name)), nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
