package updater

import (
	"archive/zip"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// GetExeDir returns the directory name of the executable
func GetExeDir() string {
	exe, _ := os.Executable()
	return filepath.Dir(exe)
}

// SHA1Hash returns the SHA1 hash as []byte
func SHA1Hash(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if nil != err {
		return []byte{}, err
	}
	defer file.Close()

	// Open a new hash interface to write to
	hash := sha1.New()

	// Copy the file into the hash interface
	_, err = io.Copy(hash, file)
	if nil != err {
		return []byte{}, err
	}

	// return []byte representation of hash
	return hash.Sum(nil), nil
}

// CreateTempDir returns a temporary directory name and error if the creation failed
func CreateTempDir() (tempDir string, err error) {
	tempDir, err = ioutil.TempDir(GetExeDir(), "updater")
	if nil != err {
		return "", err
	}
	return tempDir, nil
}

// Unzip will decompress a zip archive, moving all compressed files/folders
// to the specified output directory.
func Unzip(srcArchive string, destDir string) (root string, filenames []string, err error) {
	r, err := zip.OpenReader(srcArchive)
	if err != nil {
		err := fmt.Errorf("OpenReader() failed: %v", err)
		return "", filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(destDir, f.Name)
		filenames = append(filenames, fpath)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return "", filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		err := writeDecompressedFile(f, fpath)
		if nil != err {
			return "", filenames, err
		}
	}
	return "", filenames, nil
}

// writeDecompressedFile writes (de)compressed file to `fpath`
func writeDecompressedFile(f *zip.File, fpath string) error {
	if f.FileInfo().IsDir() {
		// Make Folder
		os.MkdirAll(fpath, os.ModePerm)
		return nil
	}

	// Make File
	if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
		return err
	}

	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		err := fmt.Errorf("OpenFile() failed: %v", err)
		return err
	}
	defer outFile.Close()

	rc, err := f.Open()
	if err != nil {
		err := fmt.Errorf("Open() failed: %v", err)
		return err
	}
	defer rc.Close()

	_, err = io.Copy(outFile, rc)
	if err != nil {
		err := fmt.Errorf("Copy() failed: %v", err)
		return err
	}

	return nil
}
