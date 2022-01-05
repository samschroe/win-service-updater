package updater

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	A_LESS_THAN_B    = -1
	A_EQUAL_TO_B     = 0
	A_GREATER_THAN_B = 1
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if !os.IsNotExist(err) {
		//file exists
		return true
	}

	// no such file or directory
	return false
}

// CompareVersions compares two versions and returns an integer that indicates
// their relationship in the sort order.
// Return a negative number if versionA is less than versionB, 0 if they're
// equal, a positive number if versionA is greater than versionB.
func CompareVersions(a string, b string) int {
	fieldsA := strings.Split(a, ".")
	fieldsB := strings.Split(b, ".")

	var shortest_len int
	if len(fieldsA) < len(fieldsB) {
		shortest_len = len(fieldsA)
	} else {
		shortest_len = len(fieldsB)
	}

	for i := 0; i < shortest_len; i++ {
		a, _ := strconv.Atoi(fieldsA[i])
		b, _ := strconv.Atoi(fieldsB[i])

		if a > b {
			return A_GREATER_THAN_B
		}
		if a < b {
			return A_LESS_THAN_B
		}
	}

	if len(fieldsA) > len(fieldsB) {
		return A_GREATER_THAN_B
	}
	if len(fieldsA) < len(fieldsB) {
		return A_LESS_THAN_B
	}

	return A_EQUAL_TO_B
}

// GetUpdateDetails finds the updtdetails.udt in a list of files extracted
// from a wyu archive. It returns a `ConfigUDT` and a list of the files to
// update.
func GetUpdateDetails(extractedFiles []string) (udt ConfigUDT, updates []string, err error) {
	udtFound := false

	for _, f := range extractedFiles {
		if filepath.Base(f) == UPDTDETAILS_UDT {
			udt, err = ParseUDT(f)
			if err != nil {
				return ConfigUDT{}, updates, err
			}
			udtFound = true
		} else {
			updates = append(updates, f)
		}
	}

	if !udtFound {
		err := fmt.Errorf("no udt file found")
		return ConfigUDT{}, updates, err
	}

	return udt, updates, nil
}

// BackupFiles copies all the files to be updated in `srcDir` to a `backupDir`
// `backupDir` is returned
func BackupFiles(updates []string, srcDir string) (backupDir string, err error) {
	backupDir, err = ioutil.TempDir("", "prefix")
	if err != nil {
		log.Fatal(err)
	}
	os.Mkdir(backupDir, 0777)

	// backup the files we are about to update
	for _, f := range updates {
		orig := path.Join(srcDir, filepath.Base(f))
		back := path.Join(backupDir, filepath.Base(f))
		_, err = CopyFile(orig, back)
		if nil != err {
			return "", err
		}
	}

	return backupDir, nil
}

func DeleteDirectory(dir string) error {
	return os.RemoveAll(dir)
}

// RollbackFiles copies all the files from `backupDir` to `dstDir`
func RollbackFiles(backupDir string, dstDir string) (err error) {
	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		orig := path.Join(backupDir, path.Base(f.Name()))
		dstFile := path.Join(dstDir, path.Base(f.Name()))
		_, err = CopyFile(orig, dstFile)
		if nil != err {
			return err
		}
	}

	return nil
}

// InstallUpdate start/stops service and moves the new files into the `installDir`
func InstallUpdate(udt ConfigUDT, srcFiles []string, installDir string) error {
	// stop services
	for _, s := range udt.ServiceToStopBeforeUpdate {
		svc := ValueToString(&s)
		service_exists, err := DoesServiceExist(svc)
		if nil != err {
			e := fmt.Errorf("failed to lookup services; %v", err)
			return e
		}
		if service_exists {
			service_running, err := IsServiceRunning(svc)
			if nil != err {
				e := fmt.Errorf("failed to get %s service status; %v", svc, err)
				return e
			}
			if service_running {
				e := StopService(svc)
				if nil != e {
					e := fmt.Errorf("failed to stop %s; %v", svc, e)
					return e
				}
			}
		}
	}

	// move the files into the "base directory"
	for _, f := range srcFiles {
		err := MoveFile(f, installDir)
		if err != nil {
			return err
		}
	}

	// start services
	for _, s := range udt.ServiceToStartAfterUpdate {
		svc := ValueToString(&s)
		e := StartService(svc)
		if nil != e {
			e := fmt.Errorf("failed to start %s; %v", svc, e)
			return e
		}
	}

	return nil
}

// MoveFile moves a `file` to `dstDir`
func MoveFile(file string, dstDir string) error {
	dst := filepath.Join(dstDir, filepath.Base(file))
	// Rename() returns *LinkError if it errs
	return os.Rename(file, dst)
}

// CopyFile copies `src` to `dst`
func CopyFile(src, dst string) (int64, error) {
	if !fileExists(src) {
		return 0, nil
	}

	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()

	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
