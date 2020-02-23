package definition

import (
	"fmt"

	"github.com/maxlaverse/image-builder/pkg/builder"
)

// NewDefinitionFromPath returns a new BuilderDef instance
func NewDefinitionFromPath(location, name string) (builder.Definition, error) {
	var err error
	var buildDef builder.Definition

	if IsSourceGit(location) {
		buildDef, err = FromGit(location, name)
	} else {
		buildDef, err = FromFileSystem(location, name)
	}

	if err != nil {
		return nil, err
	}
	if len(buildDef.GetStages()) == 0 {
		return nil, fmt.Errorf("No stages found in '%s'", location)
	}
	return buildDef, err
}
