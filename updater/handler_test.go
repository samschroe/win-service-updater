package updater

import (
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

func TestUpdateHandler(t *testing.T) {
	wycFile := "../test_files/client.1.0.1.wyc"
	wysFile := "../test_files/widgetX.1.0.1.wys"
	wyuFile := "../test_files/widgetX.1.0.1.wyu"

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

func TestUpdateHandler_InvalidWYC(t *testing.T) {
	wycFile := "../test_files/foo.wyc"
	wysFile := "../test_files/widgetX.1.0.1.wys"
	wyuFile := "../test_files/widgetX.1.0.1.wyu"

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

func TestUpdateHandler_DownloadWYS_error(t *testing.T) {
	wycFile := "../test_files/client.1.0.1.wyc"
	wysFile := "../test_files/widgetX.1.0.1.wys"
	wyuFile := "../test_files/widgetX.1.0.1.wyu"

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

func TestUpdateHandler_InvalidWYS_error(t *testing.T) {
	wycFile := "../test_files/client.1.0.1.wyc"
	// wysFile := "../test_files/widgetX.1.0.1.wys"
	wyuFile := "../test_files/widgetX.1.0.1.wyu"

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

func TestUpdateHandler_DownloadWYU_error(t *testing.T) {
	wycFile := "../test_files/client.1.0.1.wyc"
	wysFile := "../test_files/widgetX.1.0.1.wys"
	wyuFile := "../test_files/widgetX.1.0.1.wyu"

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

func TestUpdateHandler_InvalidWYU_error(t *testing.T) {
	wycFile := "../test_files/client.1.0.1.wyc"
	wysFile := "../test_files/widgetX.1.0.1.wys"
	// wyuFile := "../test_files/widgetX.1.0.1.wyu"

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

func TestUpdateHandler_NoSignedHash(t *testing.T) {
	wycFile := "../test_files/client.1.0.1.wyc"
	wysFile := "../test_files/widgetX.1.0.1.wys"
	wyuFile := "../test_files/widgetX.1.0.1.wyu"

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

func TestUpdateHandler_VerifyHash_error(t *testing.T) {
	wycFile := "../test_files/client.1.0.1.wyc"
	wysFile := "../test_files/widgetX.1.0.1.wys"
	wyuFile := "../test_files/widgetX.1.0.1.wyu"

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

func TestUpdateHandler_VerifyAdler32Checksum_error(t *testing.T) {
	wycFile := "../test_files/client.1.0.1.wyc"
	wysFile := "../test_files/widgetX.1.0.1.wys"
	wyuFile := "../test_files/widgetX.1.0.1.wyu"

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

func TestUpdateHandler_GetUpdateDetails_error(t *testing.T) {
	wycFile := "../test_files/client.1.0.1.wyc"
	wysFile := "../test_files/widgetX.1.0.1.wys"
	wyuFile := "../test_files/test.zip"

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

func TestUpdateHandler_Adler32_error(t *testing.T) {
	wycFile := "../test_files/client.1.0.1.wyc"
	wysFile := "../test_files/widgetX.1.0.1.wys"
	// wyuFile := "../test_files/widgetX.1.0.1.wyu"

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
		w.Write([]byte("not a wyu archive"))
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

func TestCheckForUpdateHandler_NoUpdate(t *testing.T) {
	wycFile := "../test_files/client.1.0.1.wyc"
	wysFile := "../test_files/widgetX.1.0.1.wys"

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

func TestCheckForUpdateHandler_ErrorBadWYCFile(t *testing.T) {
	wycFile := "../test_files/foo"
	wysFile := "../test_files/widgetX.1.0.1.wys"

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

func TestCheckForUpdateHandler_ErrorHTTP(t *testing.T) {
	// wycFile := "../test_files/client.1.0.1.wyc"
	// wysFile := "../test_files/widgetX.1.0.1.wys"

	var args Args
	f := SetupTmpLog()
	args.OutputinfoLog = f.Name()
	defer TearDown(args.OutputinfoLog)
	defer f.Close()

	info := Info{}

	exitCode, _ := CheckForUpdateHandler(info, args)
	assert.Equal(t, exitCode, EXIT_ERROR)
	assert.True(t, fileExists(args.OutputinfoLog))
}

func TestCheckForUpdateHandler_ErrorBadWYSFile(t *testing.T) {
	wycFile := "../test_files/client.1.0.1.wyc"

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
