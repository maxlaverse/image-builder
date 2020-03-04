package source

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullGitSource(t *testing.T) {
	isOfType := IsSourceGit("ssh://git@github.com/maxlaverse/image-builder.git#master:builders")
	assert.True(t, isOfType)
}

func TestGitSourceWithoutFolder(t *testing.T) {
	isOfType := IsSourceGit("ssh://git@github.com/maxlaverse/image-builder.git#master")
	assert.True(t, isOfType)
}

func TestGitSourceWithoutBranch(t *testing.T) {
	isOfType := IsSourceGit("ssh://git@github.com/maxlaverse/image-builder.git")
	assert.True(t, isOfType)
}

func TestGitSourceWithoutProtocol(t *testing.T) {
	isOfType := IsSourceGit("git@github.com/maxlaverse/image-builder.git")
	assert.True(t, isOfType)
}

func TestNonGitSource(t *testing.T) {
	isOfType := IsSourceGit("/etc/motd")
	assert.False(t, isOfType)
}
