package builder

import (
	"testing"

	"github.com/maxlaverse/image-builder/pkg/config"
	enginetest "github.com/maxlaverse/image-builder/pkg/engine/test"
	executortest "github.com/maxlaverse/image-builder/pkg/executor/test"
	"github.com/maxlaverse/image-builder/pkg/template"
	"github.com/stretchr/testify/assert"
)

func TestEmptyBuildStage(t *testing.T) {
	fakeExecutor := executortest.New()
	buildConf := config.BuildConfiguration{}
	resolver := func(string) (string, error) { return "none", nil }
	dockerfile := template.NewDockerfile([]byte{}, buildConf, "../../fixtures/empty", "../../fixtures/empty", resolver, fakeExecutor)

	stage := NewBuildStage("empty", dockerfile, []string{})

	err := stage.ComputeContentHash()

	assert.NoError(t, err)
	assert.Equal(t, "4cc05c0e", stage.ContentHash())
}

func TestBuildStage(t *testing.T) {
	fakeExecutor := executortest.New()
	buildConf := config.BuildConfiguration{}
	resolver := func(string) (string, error) { return "none", nil }
	dockerfile := template.NewDockerfile([]byte{}, buildConf, "../../fixtures/empty", "../../fixtures/empty", resolver, fakeExecutor)

	stage := NewBuildStage("empty", dockerfile, []string{})
	stage.SetImageURL("final-image")
	fakeEngine := enginetest.New()
	err := stage.Build(fakeEngine)

	assert.NoError(t, err)
	assert.Equal(t, []string{"Build(final-image)"}, fakeEngine.MethodCalls)
}
