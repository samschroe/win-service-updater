package updater

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TEST_SERVICE_NAME = `system/com.huntresslabs.test`
	PLIST_BASE_NAME   = `com.huntresslabs.test.plist`
	PLIST_PATH        = "/Library/LaunchDaemons/" + PLIST_BASE_NAME
)

// TestServiceControl_IsRunning ...
func TestServiceControl_IsRunning_not(t *testing.T) {
	running, _ := IsServiceRunning(TEST_SERVICE_NAME)
	assert.False(t, running)
}

// TestServiceControl_IsRunning ...
func TestServiceControl_IsRunning(t *testing.T) {
	if !amIAdmin() {
		t.Skip("not running as root so not testing starting a service ")
	}
	err := startTestService()
	if err != nil {
		t.Fatal(err)
	}
	defer stopTestService()

	running, _ := IsServiceRunning(TEST_SERVICE_NAME)
	assert.True(t, running)

}

func stopTestService() {
	StopService(PLIST_PATH)
	os.Remove(PLIST_PATH)
}

// startTestService ...
func startTestService() (err error) {

	outf, err := os.Create(PLIST_PATH)
	if err != nil {
		return err
	}
	defer func() {
		outf.Close()
		if err != nil {
			os.Remove(PLIST_PATH)
		}
	}()

	inf, err := os.Open("testdata/" + PLIST_BASE_NAME)
	if err != nil {
		return err
	}
	defer inf.Close()

	_, err = io.Copy(outf, inf)
	if err != nil {
		return err
	}

	outf.Close()

	err = StartService(PLIST_PATH)
	if err != nil {
		return err
	}

	return nil
}
