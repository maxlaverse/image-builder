package builder

import (
	"sort"
	"testing"

	"github.com/maxlaverse/image-builder/pkg/config"
	enginetest "github.com/maxlaverse/image-builder/pkg/engine/test"
	executortest "github.com/maxlaverse/image-builder/pkg/executor/test"
	"github.com/stretchr/testify/assert"
)

func TestUnknownStage(t *testing.T) {
	fakeEngine := enginetest.New()
	fakeExecutor := executortest.New()
	builderDef := NewDefinitionFromPath("self-reference", "../../fixtures/self-reference")
	b := NewBuild(fakeEngine, fakeExecutor, builderDef, config.BuildConfiguration{}, BuildOptions{}, "fake-target-image", "../../fixtures")
	_, err := b.PrepareStages([]string{"2"})

	assert.Error(t, err)
	assert.EqualError(t, err, "failed to read the Dockerfile template: Failed to read the Dockerfile template: open ../../fixtures/self-reference/2/Dockerfile: no such file or directory")
}

func TestSelfReference(t *testing.T) {
	fakeEngine := enginetest.New()
	fakeExecutor := executortest.New()
	builderDef := NewDefinitionFromPath("self-reference", "../../fixtures/self-reference")
	b := NewBuild(fakeEngine, fakeExecutor, builderDef, config.BuildConfiguration{}, BuildOptions{}, "fake-target-image", "../../fixtures")
	stages, err := b.PrepareStages([]string{"1"})

	assert.Error(t, err)
	assert.EqualError(t, err, `failed to render the Dockerfile template: template: dockerfile:1:7: executing "dockerfile" at <BuilderStage "1">: error calling BuilderStage: cannot replace BuilderStage('1'): stage '1' is already being built - possible loop in the stage dependencies`)
	assert.Len(t, stages, 0)
}

func TestCircularReference(t *testing.T) {
	fakeEngine := enginetest.New()
	fakeExecutor := executortest.New()
	builderDef := NewDefinitionFromPath("circular-reference", "../../fixtures/circular-reference")
	b := NewBuild(fakeEngine, fakeExecutor, builderDef, config.BuildConfiguration{}, BuildOptions{}, "fake-target-image", "../../fixtures")
	stages, err := b.PrepareStages([]string{"1"})

	assert.Error(t, err)
	assert.EqualError(t, err, `failed to render the Dockerfile template: template: dockerfile:1:7: executing "dockerfile" at <BuilderStage "2">: error calling BuilderStage: cannot replace BuilderStage('2'): failed to render the Dockerfile template: template: dockerfile:1:7: executing "dockerfile" at <BuilderStage "1">: error calling BuilderStage: cannot replace BuilderStage('1'): stage '1' is already being built - possible loop in the stage dependencies`)
	assert.Len(t, stages, 0)
}

func TestComplex(t *testing.T) {
	fakeEngine := enginetest.New()
	fakeExecutor := executortest.New()
	builderDef := NewDefinitionFromPath("complex", "../../fixtures/complex")
	b := NewBuild(fakeEngine, fakeExecutor, builderDef, config.BuildConfiguration{}, BuildOptions{}, "fake-target-image", "../../fixtures")
	stages, err := b.PrepareStages([]string{"1"})

	assert.NoError(t, err)
	if !assert.Len(t, stages, 5) {
		t.Fail()
	}
	contentHashs := []string{}
	for _, v := range stages {
		contentHashs = append(contentHashs, v.ContentHash())
	}
	sort.Strings(contentHashs)
	assert.Equal(t, []string{"233d6a91", "291f2443", "4cdb0cd2", "870c6386", "870c6386"}, contentHashs)
}
