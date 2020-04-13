package fileutils

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar"
	log "github.com/sirupsen/logrus"
)

func ListMatchingFiles(srcPath string, includePatterns []string) ([]string, error) {
	log.Tracef("Include patterns are: %v", includePatterns)

	files := []string{}
	walkRoot := getWalkRoot(srcPath, ".")
	err := filepath.Walk(walkRoot, func(filePath string, f os.FileInfo, err error) error {
		relFilePath, err := filepath.Rel(srcPath, filePath)
		if err != nil {
			return err
		}
		for _, i := range includePatterns {
			match, err := doublestar.Match(i, relFilePath)
			if err != nil {
				return err
			}
			if match {
				files = append(files, relFilePath)
				return nil
			}
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func getWalkRoot(srcPath string, include string) string {
	return strings.TrimSuffix(srcPath, string(filepath.Separator)) + string(filepath.Separator) + include
}
