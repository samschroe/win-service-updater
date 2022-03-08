package updater

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeUpdateInfo struct {
	ConfigWYS ConfigWYS
	ModifyWYS bool
	ConfigIUC ConfigIUC
	ModifyIUC bool
	Err       error
}

func (fakeier FakeUpdateInfo) ParseWYC(wycFile string) (iuc ConfigIUC, err error) {
	info := Info{}

	iuc, err = info.ParseWYC(wycFile)
	if fakeier.ModifyIUC {
		iuc.IucPublicKey = fakeier.ConfigIUC.IucPublicKey
	}

	return iuc, err
}

func (fakeier FakeUpdateInfo) ParseWYS(wysFile string, args Args) (wys ConfigWYS, err error) {
	info := Info{}

	wys, err = info.ParseWYS(wysFile, args)

	if fakeier.ModifyWYS {
		wys.FileSha1 = fakeier.ConfigWYS.FileSha1
		wys.UpdateFileAdler32 = fakeier.ConfigWYS.UpdateFileAdler32
	}

	return wys, err
}

func SetupTmpLog() *os.File {
	tmpFile, err := ioutil.TempFile("", "tmpLog")
	if err != nil {
		log.Fatal(err)
	}
	return tmpFile
}

func TearDown(f string) {
	err := os.Remove(f)
	if err != nil {
		log.Fatal(err)
	}
}

func TestHandler_UpdateHandler_InvalidWYC(t *testing.T) {
	wycFile := "./testdata/foo.wyc"
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
	assert.Equal(t, EXIT_ERROR, exitCode)
	assert.NotNil(t, err)
	assert.True(t, fileExists(args.OutputinfoLog))
}

func TestHandler_UpdateHandler_download_WYS_error(t *testing.T) {
	wycFile := "./testdata/client.1.0.1.wyc"
	wysFile := "./testdata/widgetX.1.0.1.wys"
	wyuFile := "./testdata/widgetX.1.0.1.wyu"

	// wys server
	tsWYS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
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

	finfo := FakeUpdateInfo{}
	finfo.ModifyWYS = true
	finfo.ConfigWYS.FileSha1 = []byte("invalid")

	exitCode, err := UpdateHandler(finfo, args)
	assert.Equal(t, EXIT_ERROR, exitCode)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Error downloading")
}

func TestHandler_UpdateHandler_invalid_WYS_error(t *testing.T) {
	wycFile := "./testdata/client.1.0.1.wyc"
	// wysFile := "./testdata/widgetX.1.0.1.wys"
	wyuFile := "./testdata/widgetX.1.0.1.wyu"

	// wys server
	tsWYS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not a wys file"))
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

	finfo := FakeUpdateInfo{}

	exitCode, err := UpdateHandler(finfo, args)
	assert.Equal(t, EXIT_ERROR, exitCode)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error reading wys file")
}

func TestHandler_UpdateHandler_download_WYU_error(t *testing.T) {
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
		w.WriteHeader(http.StatusForbidden)
		dat, err := ioutil.ReadFile(wyuFile)
		assert.Nil(t, err)
		w.Write(dat)
	}))
	defer tsWYU.Close()

	var args Args
	args.Cdata = wycFile
	args.Server = tsWYS.URL
	args.WYUTestServer = tsWYU.URL

	finfo := FakeUpdateInfo{}
	finfo.ModifyWYS = true
	finfo.ConfigWYS.FileSha1 = []byte("invalid")

	exitCode, err := UpdateHandler(finfo, args)
	assert.Equal(t, EXIT_ERROR, exitCode)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Error downloading")
}

func TestHandler_UpdateHandler_invalid_WYU_error(t *testing.T) {
	wycFile := "./testdata/client.1.0.1.wyc"
	wysFile := "./testdata/widgetX.1.0.1.wys"
	// wyuFile := "./testdata/widgetX.1.0.1.wyu"

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
		// dat, err := ioutil.ReadFile(wyuFile)
		// assert.Nil(t, err)
		w.Write([]byte("not a zip"))
	}))
	defer tsWYU.Close()

	var args Args
	args.Cdata = wycFile
	args.Server = tsWYS.URL
	args.WYUTestServer = tsWYU.URL

	finfo := FakeUpdateInfo{}
	finfo.ModifyIUC = true
	finfo.ConfigIUC.IucPublicKey = TLV{}

	exitCode, err := UpdateHandler(finfo, args)
	assert.Equal(t, EXIT_ERROR, exitCode)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error unzipping")
}

func TestHandler_UpdateHandler_update_not_signed(t *testing.T) {
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

	finfo := FakeUpdateInfo{}
	finfo.ModifyWYS = true
	finfo.ConfigWYS.FileSha1 = make([]byte, 0)

	exitCode, err := UpdateHandler(finfo, args)
	assert.Equal(t, EXIT_ERROR, exitCode)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "The update is not signed. All updates must be signed in order to be installed.")
}

func TestHandler_UpdateHandler_signature_verification_error(t *testing.T) {
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

	finfo := FakeUpdateInfo{}
	finfo.ModifyWYS = true
	finfo.ConfigWYS.FileSha1 = []byte("invalid")

	exitCode, err := UpdateHandler(finfo, args)
	assert.Equal(t, EXIT_ERROR, exitCode)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "crypto/rsa: verification error")
}

func TestHandler_UpdateHandler_checksum_error(t *testing.T) {
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

	finfo := FakeUpdateInfo{}
	finfo.ModifyIUC = true
	finfo.ConfigIUC.IucPublicKey = TLV{}
	finfo.ModifyWYS = true
	finfo.ConfigWYS.UpdateFileAdler32 = 1
	finfo.ConfigWYS.FileSha1 = make([]byte, 0)

	exitCode, err := UpdateHandler(finfo, args)
	assert.Equal(t, EXIT_ERROR, exitCode)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed the Adler32 validation.")
}

func TestHandler_UpdateHandler_no_updtdetails_file_error(t *testing.T) {
	wycFile := "./testdata/client.1.0.1.wyc"
	wysFile := "./testdata/widgetX.1.0.1.wys"
	wyuFile := "./testdata/test.zip"

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

	finfo := FakeUpdateInfo{}
	finfo.ModifyIUC = true
	finfo.ConfigIUC.IucPublicKey = TLV{}
	finfo.ModifyWYS = true
	a32, _ := GetAdler32(wyuFile)
	finfo.ConfigWYS.UpdateFileAdler32 = int64(a32)
	finfo.ConfigWYS.FileSha1 = make([]byte, 0)

	exitCode, err := UpdateHandler(finfo, args)
	assert.Equal(t, EXIT_ERROR, exitCode)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no udt file found")
}

func TestHandler_CheckForUpdateHandler_no_update(t *testing.T) {
	wycFile := "./testdata/client.1.0.1.wyc"
	wysFile := "./testdata/widgetX.1.0.1.wys"

	// wys server
	tsWYS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		dat, err := ioutil.ReadFile(wysFile)
		assert.Nil(t, err)
		w.Write(dat)
	}))
	defer tsWYS.Close()

	var args Args
	args.Cdata = wycFile
	args.Server = tsWYS.URL
	args.Outputinfo = true
	f := SetupTmpLog()
	args.OutputinfoLog = f.Name()
	defer TearDown(args.OutputinfoLog)
	defer f.Close()

	info := Info{}

	exitCode, _ := CheckForUpdateHandler(info, args)
	assert.Equal(t, exitCode, EXIT_NO_UPDATE)
	assert.True(t, fileExists(args.OutputinfoLog))
}

func TestHandler_CheckForUpdateHandler_invalid_WYC_file(t *testing.T) {
	wycFile := "./testdata/foo"
	wysFile := "./testdata/widgetX.1.0.1.wys"

	// wys server
	tsWYS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		dat, err := ioutil.ReadFile(wysFile)
		assert.Nil(t, err)
		w.Write(dat)
	}))
	defer tsWYS.Close()

	var args Args
	args.Cdata = wycFile
	args.Server = tsWYS.URL
	args.Outputinfo = true
	f := SetupTmpLog()
	args.OutputinfoLog = f.Name()
	defer TearDown(args.OutputinfoLog)
	defer f.Close()

	info := Info{}

	exitCode, _ := CheckForUpdateHandler(info, args)
	assert.Equal(t, exitCode, EXIT_ERROR)
	assert.True(t, fileExists(args.OutputinfoLog))
}

func TestHandler_CheckForUpdateHandler_http_error(t *testing.T) {
	wycFile := "./testdata/client.1.0.1.wyc"

	var args Args
	args.Cdata = wycFile
	args.Server = "http://foo.bar"
	f := SetupTmpLog()
	args.OutputinfoLog = f.Name()
	defer TearDown(args.OutputinfoLog)
	defer f.Close()

	info := Info{}

	exitCode, err := CheckForUpdateHandler(info, args)
	assert.NotNil(t, err)
	assert.Equal(t, exitCode, EXIT_ERROR)
	assert.True(t, fileExists(args.OutputinfoLog))
}

func TestHandler_CheckForUpdateHandler_invalid_WYS_file(t *testing.T) {
	wycFile := "./testdata/client.1.0.1.wyc"

	// wys server
	tsWYS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not a wys file"))
	}))
	defer tsWYS.Close()

	var args Args
	args.Cdata = wycFile
	args.Server = tsWYS.URL
	f := SetupTmpLog()
	args.OutputinfoLog = f.Name()
	defer TearDown(args.OutputinfoLog)
	defer f.Close()

	info := Info{}

	exitCode, _ := CheckForUpdateHandler(info, args)
	assert.Equal(t, exitCode, EXIT_ERROR)
	assert.True(t, fileExists(args.OutputinfoLog))
}

// prettyPrint is a nice to have debugging function (prints out
// structures nicely)
func prettyPrint(o interface{}) string {
	jayson, _ := json.MarshalIndent(o, "", "\t")
	return string(jayson)
}

// MarshalJSON prints out TLV structures so that you can make sense of
// them
func (t TLV) MarshalJSON() ([]byte, error) {
	type localTlv TLV
	type tlvPlus struct {
		localTlv
		ValueString string
	}

	return json.Marshal(tlvPlus{localTlv(t), string(t.Value)})
}
