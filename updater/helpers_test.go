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

func Sha256Hash(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var sum string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return sum, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := sha256.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return sum, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	sum = hex.EncodeToString(hashInBytes)
	return sum, nil
}
