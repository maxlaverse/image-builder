package source

import (
	"path"
)

// FromFilesystem simply returns the path to the builder definition on the
// local filesystem
func FromFilesystem(name, location string) (string, error) {
	return path.Join(location, name), nil
}
