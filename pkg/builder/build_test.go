package builder

import (
	"fmt"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/maxlaverse/image-builder/pkg/config"
	enginetest "github.com/maxlaverse/image-builder/pkg/engine/test"
	executortest "github.com/maxlaverse/image-builder/pkg/executor/test"
	"github.com/stretchr/testify/assert"
)

func TestPrepareUnknownStage(t *testing.T) {
	fakeEngine := enginetest.New()
	fakeExecutor := executortest.New()
	builderDef := NewDefinitionFromPath("self-reference", "../../fixtures/self-reference")
	b := NewBuild(fakeEngine, fakeExecutor, builderDef, config.BuildConfiguration{}, BuildOptions{}, "fake-target-image", "../../fixtures")
	_, err := b.PrepareStages([]string{"2"})

	assert.Error(t, err)
	assert.EqualError(t, err, "failed to read the Dockerfile template: Failed to read the Dockerfile template: open ../../fixtures/self-reference/2/Dockerfile: no such file or directory")
}

func TestPrepareSelfReference(t *testing.T) {
	fakeEngine := enginetest.New()
	fakeExecutor := executortest.New()
	builderDef := NewDefinitionFromPath("self-reference", "../../fixtures/self-reference")
	b := NewBuild(fakeEngine, fakeExecutor, builderDef, config.BuildConfiguration{}, BuildOptions{}, "fake-target-image", "../../fixtures")
	stages, err := b.PrepareStages([]string{"1"})

	assert.Error(t, err)
	assert.EqualError(t, err, `failed to render the Dockerfile template: template: dockerfile:1:7: executing "dockerfile" at <BuilderStage "1">: error calling BuilderStage: cannot replace BuilderStage('1'): stage '1' is already being built - possible loop in the stage dependencies`)
	assert.Len(t, stages, 0)
}

func TestPrepareCircularReference(t *testing.T) {
	fakeEngine := enginetest.New()
	fakeExecutor := executortest.New()
	builderDef := NewDefinitionFromPath("circular-reference", "../../fixtures/circular-reference")
	b := NewBuild(fakeEngine, fakeExecutor, builderDef, config.BuildConfiguration{}, BuildOptions{}, "fake-target-image", "../../fixtures")
	stages, err := b.PrepareStages([]string{"1"})

	assert.Error(t, err)
	assert.EqualError(t, err, `failed to render the Dockerfile template: template: dockerfile:1:7: executing "dockerfile" at <BuilderStage "2">: error calling BuilderStage: cannot replace BuilderStage('2'): failed to render the Dockerfile template: template: dockerfile:1:7: executing "dockerfile" at <BuilderStage "1">: error calling BuilderStage: cannot replace BuilderStage('1'): stage '1' is already being built - possible loop in the stage dependencies`)
	assert.Len(t, stages, 0)
}

func TestPrepareComplex(t *testing.T) {
	fakeEngine := enginetest.New()
	fakeExecutor := executortest.New()
	builderDef := NewDefinitionFromPath("complex", "../../fixtures/complex")
	b := NewBuild(fakeEngine, fakeExecutor, builderDef, config.BuildConfiguration{}, BuildOptions{}, "fake-target-image", "../../fixtures/empty")
	stages, err := b.PrepareStages([]string{"1"})

	assert.NoError(t, err)
	if !assert.Len(t, stages, 5) {
		t.Fail()
	}
	assert.Equal(t, []string{"1=add0b833", "2=b8734280", "3=f829d8c6", "4=9a7eedee", "5=9a7eedee"}, stagesToHashes(stages))
}

func TestBuildConcurrently(t *testing.T) {
	fakeEngine := enginetest.NewWithCallbacks(delayInOrder(t, map[string]time.Duration{
		"fake-target-image:final-1bb557c5":        0,
		"fake-target-image:parallel-1-1-ebe68df7": time.Duration(0),
		"fake-target-image:parallel-1-2-01123178": time.Duration(10) * time.Millisecond,
		"fake-target-image:parallel-2-1-ebe68df7": time.Duration(5) * time.Millisecond,
		"fake-target-image:parallel-2-2-768ce388": time.Duration(0),
	}))
	fakeExecutor := executortest.New()
	builderDef := NewDefinitionFromPath("concurrency", "../../fixtures/concurrency")

	for i := 0; i < 10; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			b := NewBuild(fakeEngine, fakeExecutor, builderDef, config.BuildConfiguration{}, BuildOptions{BuildConcurrency: 2}, "fake-target-image", "../../fixtures/empty")

			stages, err := b.BuildStages([]string{"final"})
			assert.NoError(t, err)
			if !assert.Len(t, stages, 5) {
				t.Fail()
			}
			assert.Equal(t, []string{"final=1bb557c5", "parallel-1-1=ebe68df7", "parallel-1-2=01123178", "parallel-2-1=ebe68df7", "parallel-2-2=768ce388"}, stagesToHashes(stages))
			assert.Equal(t, "Build(fake-target-image:parallel-2-2-768ce388)", fakeEngine.MethodCalls[2])
			assert.Equal(t, "Build(fake-target-image:parallel-1-2-01123178)", fakeEngine.MethodCalls[3])
			assert.Equal(t, "Build(fake-target-image:final-1bb557c5)", fakeEngine.MethodCalls[4])
		})
	}
}

func stagesToHashes(stages []BuildStage) []string {
	contentHashs := []string{}
	for _, v := range stages {
		contentHashs = append(contentHashs, fmt.Sprintf("%s=%s", v.Name(), v.ContentHash()))
	}
	sort.Strings(contentHashs)
	return contentHashs
}

func delayInOrder(t *testing.T, delays map[string]time.Duration) func(string) {
	return func(image string) {
		v, ok := delays[image]
		if !ok {
			t.Fatalf("The image '%s' had no delay defined", image)
		}
		time.Sleep(v)
	}
}
