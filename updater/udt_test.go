package updater

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUDT(t *testing.T) {
	tmpfile := GenerateTempFile()
	defer os.Remove(tmpfile)

	orig := "../test_files/updtdetails.udt"
	udt, err := ParseUDT(orig)
	assert.Nil(t, err)

	err = WriteUDT(udt, tmpfile)
	assert.Nil(t, err)

	origHash, err := Sha256Hash(orig)
	assert.Nil(t, err)

	newHash, err := Sha256Hash(tmpfile)
	assert.Nil(t, err)
	assert.Equal(t, origHash, newHash)
}

func TestUDT_ParseUDT_error(t *testing.T) {
	tmpfile := GenerateTempFile()
	defer os.Remove(tmpfile)

	_, err := ParseUDT(tmpfile)
	assert.NotNil(t, err)

	_, err = ParseUDT("")
	assert.NotNil(t, err)

	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, []byte(UPDTDETAILS_HEADER))
	_ = binary.Write(&buf, binary.BigEndian, uint8(0xfe))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(1))
	ioutil.WriteFile(tmpfile, buf.Bytes(), 6444)
	_, err = ParseUDT(tmpfile)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
