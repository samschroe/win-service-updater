package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArgs(t *testing.T) {
	argv := []string{"win_service_updater.exe", "/noerr", "-foo"}
	args, err := ParseArgs(argv)
	assert.NotNil(t, err)

	argv = []string{"win_service_updater.exe", "/noerr", "-logfile=foo"}
	args, err = ParseArgs(argv)
	assert.Nil(t, err)
	assert.True(t, args.Noerr)
	assert.Equal(t, args.Logfile, "foo")

	argv = []string{"win_service_updater.exe", "/justcheck", "/outputinfo=foo"}
	args, err = ParseArgs(argv)
	assert.Nil(t, err)
	assert.True(t, args.Justcheck)
	assert.True(t, args.Outputinfo)
	assert.Equal(t, args.OutputinfoLog, "foo")

	argv = []string{"win_service_updater.exe", "/fromservice", "/quickcheck"}
	args, err = ParseArgs(argv)
	assert.Nil(t, err)
	assert.True(t, args.Fromservice)
	assert.True(t, args.Quickcheck)
}
