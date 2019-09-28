package updater

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFile_Unzip(t *testing.T) {
	tempDir, err := CreateTempDir()
	assert.Nil(t, err)
	defer os.RemoveAll(tempDir)

	zipFile := "../test_files/test.zip"
	_, files, err := Unzip(zipFile, tempDir)
	assert.Nil(t, err)
	// there are 4 files in the test.zip
	assert.Equal(t, 4, len(files))
}

func TestFile_Sha1Hash(t *testing.T) {
	_, err := SHA1Hash("")
	assert.NotNil(t, err)
}
