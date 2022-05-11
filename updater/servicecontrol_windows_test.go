package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/windows/svc"
)

const (
	TEST_SERVICE_NAME = `PlugPlay`
)

func TestServiceControl(t *testing.T) {
	// Verify the service is Started
	e := startTestService()
	assert.NoError(t, e)

	status, e := GetServiceState(TEST_SERVICE_NAME)
	assert.True(t, (status == svc.Running))
	assert.NoError(t, e)

	// Verify service reports stopped
	e = stopTestService()
	assert.NoError(t, e)

	status, e = GetServiceState(TEST_SERVICE_NAME)
	assert.True(t, (status == svc.Stopped))
	assert.NoError(t, e)

	// Verify service will start
	e = startTestService()
	assert.NoError(t, e)

	status, e = GetServiceState(TEST_SERVICE_NAME)
	assert.True(t, (status == svc.Running))
	assert.NoError(t, e)
}

func stopTestService() error {
	return StopService(TEST_SERVICE_NAME)
}

// startTestService ...
func startTestService() error {
	return StartService(TEST_SERVICE_NAME)
}
