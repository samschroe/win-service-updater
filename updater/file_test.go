package updater

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
)

func TestFile_Unzip(t *testing.T) {
	tempDir, err := CreateTempDir()
	assert.Nil(t, err)
	defer os.RemoveAll(tempDir)

	zipFile := "./testdata/test.zip"
	_, files, err := Unzip(zipFile, tempDir)
	assert.Nil(t, err)
	// there are 4 files in the test.zip
	assert.Equal(t, 4, len(files))
}

func Test_GenerateSHA1HashFromReader_ReturnsErrorWhenReaderFails(t *testing.T) {
	reader := iotest.ErrReader(errors.New("Me no know how to readz!"))

	hash, err := GenerateSHA1HashFromReader(reader)
	assert.NotNil(t, err)
	assert.Empty(t, hash)
}

func Test_GenerateSHA1HashFromReader_ReturnsHashFromReader(t *testing.T) {
	reader := strings.NewReader("KNOWN VALUE")
	expectedHash := []byte{0x98, 0x39, 0x87, 0x99, 0x8a, 0xf6, 0xe1, 0x28, 0xb5, 0xf8, 0xd8, 0x5b, 0xc7, 0x63, 0x42, 0x78, 0xf7, 0x47, 0x85, 0xde}

	actualHash, err := GenerateSHA1HashFromReader(reader)
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(expectedHash, actualHash))
}

func Test_GenerateSHA1HashFromFilePath_ReturnsErrorWhenFailToOpenFile(t *testing.T) {
	hash, err := GenerateSHA1HashFromFilePath("this is definitely not a file path that exists :-P")
	assert.NotNil(t, err)
	assert.Empty(t, hash)
}

func Test_GenerateSHA1HashFromFilePath_ReturnsHashOfFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "i-am-a-file")
	assert.NoError(t, os.WriteFile(filePath, []byte("KNOWN VALUE"), 0600))

	expectedHash := []byte{0x98, 0x39, 0x87, 0x99, 0x8a, 0xf6, 0xe1, 0x28, 0xb5, 0xf8, 0xd8, 0x5b, 0xc7, 0x63, 0x42, 0x78, 0xf7, 0x47, 0x85, 0xde}

	actualHash, err := GenerateSHA1HashFromFilePath(filePath)
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(expectedHash, actualHash))
}
