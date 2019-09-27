package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdler32(t *testing.T) {
	f := GenerateTempFile()
	adler32, err := GetAdler32(f)
	assert.Nil(t, err)

	b := VerifyAdler32Checksum(int64(adler32), "")
	assert.False(t, b)

	b = VerifyAdler32Checksum(int64(adler32), f)
	assert.True(t, b)

	_, err = GetAdler32("")
	assert.NotNil(t, err)
}
