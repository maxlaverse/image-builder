package template

import (
	"github.com/maxlaverse/image-builder/pkg/config"
)

// BuildData holds all the information required for rendering
// a Dockerfile template
type BuildData struct {
	build        config.BuildConfiguration
	deps         map[string]struct{}
	images       map[string]string
	localContext string
}

// NewMinimalBuildData returns an instance of BuildData with the build configuration
// and local context set
func NewMinimalBuildData(buildConf config.BuildConfiguration, localContext string) BuildData {
	return BuildData{
		build:        buildConf,
		deps:         map[string]struct{}{},
		images:       map[string]string{},
		localContext: localContext,
	}
}

// NewBuildData returns an instance of BuildData with the build configuration, the local context
// and a list of available stage images
func NewBuildData(buildConf config.BuildConfiguration, localContext string, stageImages map[string]string) BuildData {
	return BuildData{
		build:        buildConf,
		deps:         map[string]struct{}{},
		images:       stageImages,
		localContext: localContext,
	}
}

func (t *BuildData) dependencies() []string {
	deps := []string{}
	for dep := range t.deps {
		deps = append(deps, dep)
	}
	return deps
}
