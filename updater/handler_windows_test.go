package updater

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_UpdateHandler(t *testing.T) {
	wycFile := "./testdata/client.1.0.1.wyc"
	wysFile := "./testdata/widgetX.1.0.1.wys"
	wyuFile := "./testdata/widgetX.1.0.1.wyu"

	// wys server
	tsWYS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		dat, err := ioutil.ReadFile(wysFile)
		assert.Nil(t, err)
		w.Write(dat)
	}))
	defer tsWYS.Close()

	// wys server
	tsWYU := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		dat, err := ioutil.ReadFile(wyuFile)
		assert.Nil(t, err)
		w.Write(dat)
	}))
	defer tsWYU.Close()

	var args Args
	args.Cdata = wycFile
	args.Server = tsWYS.URL
	args.WYUTestServer = tsWYU.URL
	args.Outputinfo = true
	f := SetupTmpLog()
	args.OutputinfoLog = f.Name()
	defer TearDown(args.OutputinfoLog)
	defer f.Close()

	info := Info{}

	exitCode, err := UpdateHandler(info, args)
	assert.Equal(t, EXIT_NO_UPDATE, exitCode)
	assert.Nil(t, err)
	assert.True(t, fileExists(args.OutputinfoLog))
}
