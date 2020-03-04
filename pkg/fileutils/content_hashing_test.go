package fileutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNegation(t *testing.T) {
	pm, err := NewPatternMatcher([]string{"*", "!Gemfile*", ".dockerignore", "build.yaml"})
	assert.NoError(t, err)

	skip, err := pm.Matches("Gemfile.lock")
	assert.NoError(t, err)
	assert.False(t, skip)
}

func TestExclusin(t *testing.T) {
	pm, err := NewPatternMatcher([]string{"*", "!Gemfile*", ".dockerignore", "build.yaml"})
	assert.NoError(t, err)

	skip, err := pm.Matches("Gemnotfile.lock")
	assert.NoError(t, err)
	assert.True(t, skip)
}
