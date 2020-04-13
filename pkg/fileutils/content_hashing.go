package fileutils

import (
	"encoding/hex"
	"hash/crc32"
	"io"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

func ContentHashing(basePath string, files []string, extraContent string) (string, error) {
	//Initialize an empty return string now in case an error has to be returned
	var returnCRC32String string

	//Open the fhe file located at the given path and check for errors

	//Create the table with the given polynomial
	tablePolynomial := crc32.MakeTable(0xedb88320)

	//Open a new hash interface to write the file to
	hash := crc32.New(tablePolynomial)

	//Copy the file in the interface

	for _, filePath := range files {
		fi, err := os.Stat(path.Join(basePath, filePath))
		if err != nil {
			log.Errorf("Could not stat: '%s'", path.Join(basePath, filePath))
			continue
			return "", err
		}

		if fi.IsDir() {
			continue
		}

		err = func() error {
			file, err := os.Open(path.Join(basePath, filePath))
			if err != nil {
				log.Errorf("Could not open: '%s'", path.Join(basePath, filePath))
				return err
			}

			//Tell the program to close the file when the function returns
			defer file.Close()

			// TODO: Include permissions
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
