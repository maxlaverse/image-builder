package fileutils

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

func ListMatchingFiles(srcPath string, ignorePatterns []string) ([]string, error) {
	files := []string{}

	log.Tracef("Ignore patterns are: %v", ignorePatterns)
	pm, err := NewPatternMatcher(ignorePatterns)
	if err != nil {
		return files, err
	}

	include := "."
	walkRoot := getWalkRoot(srcPath, include)
	err = filepath.Walk(walkRoot, func(filePath string, f os.FileInfo, err error) error {
		relFilePath, err := filepath.Rel(srcPath, filePath)
		if err != nil {
			return err
		}
		skip := false
		if include != relFilePath {
			skip, err = pm.Matches(f.Name())
			if err != nil {
				return err
			}
		}
		if skip {
			log.Tracef("Skipping excluded path: %s", f.Name())

			if !f.IsDir() {
				return nil
			}

			if !pm.Exclusions() {
				return filepath.SkipDir
			}

			dirSlash := relFilePath + string(filepath.Separator)
			for _, pat := range pm.Patterns() {
				if !pat.Exclusion() {
					continue
				}
				if strings.HasPrefix(pat.String()+string(filepath.Separator), dirSlash) {
					// found a match - so can't skip this dir
					return nil
				}
			}

			// No matching exclusion dir so just skip dir
			log.Tracef("Skipping whole directory: %s", f.Name())
			return filepath.SkipDir
		}
		log.Tracef("Including path in context: %s", relFilePath)
		files = append(files, relFilePath)
		return nil
	})
	sort.Strings(files)
	return files, err
}

func getWalkRoot(srcPath string, include string) string {
	return strings.TrimSuffix(srcPath, string(filepath.Separator)) + string(filepath.Separator) + include
}
