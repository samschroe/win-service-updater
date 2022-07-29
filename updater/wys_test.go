package updater

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWYS_ParseWYS(t *testing.T) {
	info := Info{}
	var args Args
	wys, err := info.ParseWYS("./testdata/widgetX.1.0.1.wys", args)
	assert.Nil(t, err)
	assert.Contains(t, wys.UpdateFileSite[0], "127.0.0.1")
}

func TestWYS_ReadWYSTLV(t *testing.T) {
	r := bytes.NewReader([]byte{})
	tlv := ReadWYSTLV(r)
	assert.Nil(t, tlv)
}

func TestWYS_getWyuFile(t *testing.T) {
	info := Info{}
	var args Args
	// the fake wys file
	wys, err := info.ParseWYS("./testdata/widgetX.1.0.1.wys", args)
	assert.NoError(t, err)

	// the fake wyuFile
	const wyuFile = "./testdata/fake-wyu.wyu"

	// count the number of times the download endpoint is hit
	downloadCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read the wyu file and return the data
		data, _ := ioutil.ReadFile(wyuFile)
		w.Write(data)
		// increment the download count
		downloadCount = downloadCount + 1
	}))
	defer ts.Close()

	// create a temp dir to play in
	baseDir, _ := ioutil.TempDir("", "test-getWyuFile")
	defer os.RemoveAll(baseDir)

	// name the cached wyu location and the download location
	lastWyuFilePath = filepath.Join(baseDir, "last-download-wyu")
	downloadLoc := filepath.Join(baseDir, "wyu-download")

	// diddle the wys struct URLs to point to our test server
	wys.UpdateFileSite = []string{ts.URL}
	// diddle the wys struct to have the correct adler32
	// calculation
	// get the adler32 for the wyuFile
	adler32, _ := GetAdler32(wyuFile)
	wys.UpdateFileAdler32 = int64(adler32)

	// get the wyu file. We expect that download count to
	// increment
	err = wys.getWyuFile(args, downloadLoc)
	assert.NoError(t, err)
	assert.Equal(t, 1, downloadCount)

	// remove the downloadLoc file just because
	assert.NoError(t, os.Remove(downloadLoc))

	// get the wyu file again. We expect to use the locally cached
	// version and *not* increment the download count
	err = wys.getWyuFile(args, downloadLoc)
	assert.NoError(t, err)
	assert.Equal(t, 1, downloadCount)

	// we expect the downloadLoc to exist (again)
	_, err = os.Stat(downloadLoc)
	assert.NoError(t, err)
}
