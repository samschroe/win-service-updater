package updater

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func appendFiles(filename string, zipw *zip.Writer) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Failed to open %s: %s", filename, err)
	}
	defer file.Close()

	wr, err := zipw.Create(path.Base(filename))
	if err != nil {
		msg := "Failed to create entry for %s in zip file: %s"
		return fmt.Errorf(msg, filename, err)
	}

	if _, err := io.Copy(wr, file); err != nil {
		return fmt.Errorf("Failed to write %s to zip: %s", filename, err)
	}

	return nil
}

func Zip(archive string, files []string) {
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	file, err := os.OpenFile(archive, flags, 0644)
	if err != nil {
		log.Fatalf("Failed to open zip for writing: %s", err)
	}
	defer file.Close()

	zipw := zip.NewWriter(file)
	defer zipw.Close()

	for _, filename := range files {
		if err := appendFiles(filename, zipw); err != nil {
			log.Fatalf("Failed to add file %s to zip: %s", filename, err)
		}
	}
}

func TestWYC(t *testing.T) {
	origFile := "../test_files/client.1.0.0.wyc"

	info := Info{}
	wyc, err := info.ParseWYC(origFile)
	assert.Nil(t, err)
	assert.Equal(t, wyc.IucServerFileSite[0].Value, []byte("http://127.0.0.1/updates/wyserver.wys?key=%urlargs%"))
}

func TestWYC_GetWYSURLs(t *testing.T) {
	origFile := "../test_files/client.1.0.0.wyc"

	args := Args{}

	info := Info{}
	wyc, err := info.ParseWYC(origFile)
	assert.Nil(t, err)
	urls := wyc.GetWYSURLs(args)
	assert.Equal(t, 1, len(urls))
}

func TestWYC_UpdateWYCWithNewVersion(t *testing.T) {
	origFile := "../test_files/client.1.0.0.wyc"

	info := Info{}
	wyc, err := info.ParseWYC(origFile)
	assert.Nil(t, err)

	new, err := UpdateWYCWithNewVersionNumber(wyc, origFile, "1.2.3.4")
	assert.Nil(t, err)
	// fmt.Println(new)

	newConfig, err := info.ParseWYC(new)
	assert.Nil(t, err)
	assert.Equal(t, string(newConfig.IucInstalledVersion.Value), "1.2.3.4")
}

func TestWYC_WriteIUC(t *testing.T) {
	origClientWYC := "../test_files/client.1.0.0.wyc"

	// create a new uiclient.iuc and compare it to the one in the archive
	info := Info{}
	wyc, err := info.ParseWYC(origClientWYC)
	assert.Nil(t, err)

	tmpIUC := GenerateTempFile()
	defer os.Remove(tmpIUC)

	err = WriteIUC(wyc, tmpIUC)
	assert.Nil(t, err)

	tmpDir := GetTempDir()
	defer os.RemoveAll(tmpDir)

	newHash, err := Sha256Hash(tmpIUC)
	assert.Nil(t, err)

	found := false
	_, files, err := Unzip(origClientWYC, tmpDir)
	for _, f := range files {
		// fmt.Println(f)
		if filepath.Base(f) == IUCLIENT_IUC {
			origHash, err := Sha256Hash(f)
			assert.Nil(t, err)
			assert.Equal(t, origHash, newHash)
			found = true
		}
	}
	assert.True(t, found)
}
