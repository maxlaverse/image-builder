package fileutils

import (
	"encoding/hex"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

func getWalkRoot(srcPath string, include string) string {
	return strings.TrimSuffix(srcPath, string(filepath.Separator)) + string(filepath.Separator) + include
}

func ContentHashing(extraContent string, srcPath string, ignorePatterns []string) (string, error) {
	log.Debugf("Content Hashing Context is: %v", srcPath)
	log.Debugf("Ignore patterns are: %v", ignorePatterns)
	pm, err := NewPatternMatcher(ignorePatterns)
	if err != nil {
		return "", err
	}
	i := 0
	files := []string{}
	include := "."
	walkRoot := getWalkRoot(srcPath, include)
	err = filepath.Walk(walkRoot, func(filePath string, f os.FileInfo, err error) error {
		relFilePath, err := filepath.Rel(srcPath, filePath)
		if err != nil {
			return err
		}
		skip, err := pm.Matches(f.Name())
		if err != nil {
			return err
		}
		if skip {
			i = i + 1

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
			return filepath.SkipDir
		}
		log.Tracef("Including path in context: %s", relFilePath)
		files = append(files, relFilePath)
		return nil
	})
	if err != nil {
		log.Errorf("Could not walk: '%v'", err)
		return "", err
	}

	sort.Strings(files)
	//Initialize an empty return string now in case an error has to be returned
	var returnCRC32String string

	//Open the fhe file located at the given path and check for errors

	//Create the table with the given polynomial
	tablePolynomial := crc32.MakeTable(0xedb88320)

	//Open a new hash interface to write the file to
	hash := crc32.New(tablePolynomial)

	//Copy the file in the interface

	for _, filePath := range files {
		fi, err := os.Stat(srcPath + "/" + filePath)
		if err != nil {
			log.Errorf("Could not stat: '%s'", srcPath+"/"+filePath)
			continue
			return "", err
		}

		if fi.IsDir() {
			continue
		}

		err = func() error {
			file, err := os.Open(srcPath + "/" + filePath)
			if err != nil {
				log.Errorf("Could not open: '%s'", srcPath+"/"+filePath)
				return err
			}

			//Tell the program to close the file when the function returns
			defer file.Close()

			if _, err := io.Copy(hash, file); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return "", err
		}
	}

	io.Copy(hash, strings.NewReader(strings.Join(files, "\n")))
	io.Copy(hash, strings.NewReader(extraContent))

	//Generate the hash
	hashInBytes := hash.Sum(nil)[:]

	//Encode the hash to a string
	returnCRC32String = hex.EncodeToString(hashInBytes)

	//Return the output
	return returnCRC32String, nil
}
