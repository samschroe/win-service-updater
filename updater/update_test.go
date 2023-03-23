package updater

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdate_CompareVersions(t *testing.T) {
	type versionTest struct {
		a        string
		b        string
		expected int
	}

	var versionTests = []versionTest{
		{"0.5.2", "0.6.2", A_LESS_THAN_B},
		{"0.5.2", "0.5.2", A_EQUAL_TO_B},
		{"2.2.2", "2.2.2.1", A_LESS_THAN_B},
		{"3.3.3.1", "3.3.3", A_GREATER_THAN_B},
		{"1.0.0.1", "1.0.0.2", A_LESS_THAN_B},
		{"100.0.0.1", "200.0.0.2", A_LESS_THAN_B},
		{"0.0.0.5", "0.0.0.4", A_GREATER_THAN_B},
		{"10000.0.0.1", "20000.0.0.2", A_LESS_THAN_B},
	}

	for _, tt := range versionTests {
		actual := CompareVersions(tt.a, tt.b)
		assert.Equal(t, tt.expected, actual, fmt.Sprintf("a = %s; b = %s", tt.a, tt.b))
	}
}

func Test_GenerateCandidateUpdateRequest_FailsToParseWycFile(t *testing.T) {
	args := Args{Cdata: "not a real path"}
	wyFileParser := FakeUpdateInfo{}

	req, err := NewCandidateUpdateRequest(args, wyFileParser)
	assert.Zero(t, req)
	assert.NotNil(t, err)
}

func Test_GenerateCandidateUpdateRequest_FailsToDownloadWysFile(t *testing.T) {
	wycFilePath := "./testdata/client.1.0.1.wyc"

	// wys server
	tsWYS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer tsWYS.Close()

	args := Args{
		Cdata:         wycFilePath,
		WYSTestServer: tsWYS.URL,
	}
	wyFileParser := FakeUpdateInfo{}

	req, err := NewCandidateUpdateRequest(args, wyFileParser)
	assert.Zero(t, req)
	assert.NotNil(t, err)
}

func Test_GenerateCandidateUpdateRequest_FailsToParseWysFile(t *testing.T) {
	wycFilePath := "./testdata/client.1.0.1.wyc"

	// wys server
	tsWYS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// we'll write garbage back for the WYS file, so the HTTP request will succeed but parsing will fail.
		w.Write([]byte{0xDE, 0xAD, 0xBE, 0xEF})
	}))
	defer tsWYS.Close()

	args := Args{
		Cdata:         wycFilePath,
		WYSTestServer: tsWYS.URL,
	}
	wyFileParser := FakeUpdateInfo{}

	req, err := NewCandidateUpdateRequest(args, wyFileParser)
	assert.Zero(t, req)
	assert.NotNil(t, err)
}

func Test_GenerateCandidateUpdateRequest_CandidateWysFileMatchesInstallFailedSentinel(t *testing.T) {
	wycFilePath := "./testdata/client.1.0.1.wyc"

	wysFilePath := "./testdata/widgetX.1.0.1.wys"
	wysFileContents, err := os.ReadFile(wysFilePath)
	assert.Nil(t, err)

	// this logic can go any where before GenerateCandidateUpdateRequest() executes but I'll drop in
	// here as assuming reading of the WYS file was successful, we'll write the same file to disk
	// as the install failed sentinel
	sentinelFilePath := filepath.Join(GetExeDir(), INSTALL_FAILED_SENTINAL_WYS_FILE_NAME)
	err = os.WriteFile(sentinelFilePath, wysFileContents, 0600)
	assert.Nil(t, err)
	defer os.Remove(sentinelFilePath)

	// wys server
	tsWYS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(wysFileContents)
	}))
	defer tsWYS.Close()

	args := Args{
		Cdata:         wycFilePath,
		WYSTestServer: tsWYS.URL,
	}
	wyFileParser := FakeUpdateInfo{}

	req, err := NewCandidateUpdateRequest(args, wyFileParser)
	assert.Zero(t, req)
	assert.NotNil(t, err)
}

func Test_GenerateCandidateUpdateRequest_CandidateWysFileDoesNotMatchInstallFailedSentinel(t *testing.T) {
	wycFilePath := "./testdata/client.1.0.1.wyc"

	wysFilePath := "./testdata/widgetX.1.0.1.wys"
	wysFileContents, err := os.ReadFile(wysFilePath)
	assert.Nil(t, err)

	// this logic can go any where before GenerateCandidateUpdateRequest() executes but I'll drop in
	// here as assuming reading of the WYS file was successful, we'll write the same file to disk
	// as the install failed sentinel
	sentinelFilePath := filepath.Join(GetExeDir(), INSTALL_FAILED_SENTINAL_WYS_FILE_NAME)

	// we specifically write one byte different than the "candidate" WYS file to make the sentinel not match
	sentinelFileContents := bytes.Clone(wysFileContents)
	sentinelFileContents[0] = ^sentinelFileContents[0]
	err = os.WriteFile(sentinelFilePath, sentinelFileContents, 0600)
	assert.Nil(t, err)
	defer os.Remove(sentinelFilePath)

	// wys server
	tsWYS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(wysFileContents)
	}))
	defer tsWYS.Close()

	args := Args{
		Cdata:         wycFilePath,
		WYSTestServer: tsWYS.URL,
	}
	wyFileParser := FakeUpdateInfo{}

	req, err := NewCandidateUpdateRequest(args, wyFileParser)
	assert.NotEmpty(t, req.CandidateWysFileContent)
	assert.NotZero(t, req.ConfigIUC)
	assert.NotZero(t, req.ConfigWYS)
	assert.Nil(t, err)
}

func Test_GenerateCandidateUpdateRequest_InstallFailedSentinelDoesNotExist(t *testing.T) {
	wycFilePath := "./testdata/client.1.0.1.wyc"

	// wys server
	tsWYS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		wysFilePath := "./testdata/widgetX.1.0.1.wys"
		dat, err := os.ReadFile(wysFilePath)
		assert.Nil(t, err)
		w.Write(dat)
	}))
	defer tsWYS.Close()

	args := Args{
		Cdata:         wycFilePath,
		WYSTestServer: tsWYS.URL,
	}
	wyFileParser := FakeUpdateInfo{}

	assert.False(t, fileExists(filepath.Join(GetExeDir(), INSTALL_FAILED_SENTINAL_WYS_FILE_NAME)))

	req, err := NewCandidateUpdateRequest(args, wyFileParser)
	assert.NotEmpty(t, req.CandidateWysFileContent)
	assert.NotZero(t, req.ConfigIUC)
	assert.NotZero(t, req.ConfigWYS)
	assert.Nil(t, err)
}
