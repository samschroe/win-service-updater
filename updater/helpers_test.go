package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func GenerateTempFile() string {
	f, err := ioutil.TempFile("", "testing")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write([]byte("some test data"))

	return f.Name()
}

func GetTempDir() (tmpDir string) {
	tmpDir, err := ioutil.TempDir("", "prefix")
	if err != nil {
		log.Fatal(err)
	}
	return tmpDir
}

func GetSHA256(filePath string) (string, error) {
	var sum string

	file, err := os.Open(filePath)
	if nil != err {
		return sum, err
	}
	defer file.Close()

	hash := sha256.New()

	// Copy the file into the hash interface
	_, err = io.Copy(hash, file)
	if nil != err {
		return sum, err
	}

	// Convert bytes to string
	sum = hex.EncodeToString(hash.Sum(nil)[:16])
	return sum, nil
}
