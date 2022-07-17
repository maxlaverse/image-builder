package fileutils

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar"
	log "github.com/sirupsen/logrus"
)

var (
	matchCache    = map[string][]string{}
	filelistCache = map[string][]string{}
)

func cacheKey(srcPath string, includePatterns []string) string {
	return fmt.Sprintf("%s:%v", srcPath, includePatterns)
}

func ListMatchingFiles(srcPath string, includePatterns []string) ([]string, error) {
	log.Tracef("Include patterns are: %v", includePatterns)

	if v, ok := matchCache[cacheKey(srcPath, includePatterns)]; ok {
		log.Tracef("Found %d matching files in cache", len(v))
		return v, nil
	}

	files := []string{}
	var err error
	if v, ok := filelistCache[srcPath]; ok {
		log.Tracef("Found a list of %d files in cache for %s", len(v), srcPath)
		for _, filePath := range v {
			for _, i := range includePatterns {
				match, err := doublestar.Match(i, filePath)
				if err != nil {
					return nil, err
				}
				if match {
					files = append(files, filePath)
					break
				}
			}
		}
	} else {
		if filelistCache[srcPath] == nil {
			filelistCache[srcPath] = []string{}
		}
		walkRoot := getWalkRoot(srcPath, ".")
		err = filepath.WalkDir(walkRoot, func(path string, d fs.DirEntry, _ error) error {
			relFilePath, err := filepath.Rel(srcPath, path)
			if err != nil {
				return err
			}

			if relFilePath == "." {
				return nil
			}

			filelistCache[srcPath] = append(filelistCache[srcPath], relFilePath)
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
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)

	log.Tracef("Found %d files that matched", len(files))
	matchCache[cacheKey(srcPath, includePatterns)] = files
	return files, nil
}

func getWalkRoot(srcPath string, include string) string {
	return strings.TrimSuffix(srcPath, string(filepath.Separator)) + string(filepath.Separator) + include
}
