package fileutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListMatchingFilesPopulatesCache(t *testing.T) {
	resetCache()

	list, err := ListMatchingFiles("../../fixtures/folder-listing", []string{"some-f*"})

	assert.NoError(t, err)
	assert.Equal(t, []string{"some-file"}, list)
	assert.Equal(t, matchCache, map[string][]string{"../../fixtures/folder-listing:[some-f*]": list})
}

func TestListMatchingFilesDontIncludeRoot(t *testing.T) {
	resetCache()

	list, err := ListMatchingFiles("../../fixtures/folder-listing", []string{"**"})

	assert.NoError(t, err)
	assert.NotContains(t, list, ".")
}

func resetCache() {
	matchCache = map[string][]string{}
	filelistCache = map[string][]string{}
}
