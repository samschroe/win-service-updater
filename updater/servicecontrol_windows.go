//go:build windows
// +build windows

package updater

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func DoesServiceExist(serviceName string) (bool, error) {
	m, err := mgr.Connect()
	if err != nil {
		return false, err
	}
	defer m.Disconnect()

	services, err := m.ListServices()
	if err != nil {
		return false, err
	}
	for _, name := range services {
		if strings.ToLower(serviceName) == strings.ToLower(name) {
			return true, nil
		}
	}
	return false, nil
}

func GetServiceState(serviceName string) (state svc.State, e error) {
	state = svc.Stopped

	// open service manager, requires admin
	m, err := mgr.Connect()
	if err != nil {
		return state, err
	}
	defer m.Disconnect()

	// open the service
	s, err := m.OpenService(serviceName)
	if err != nil {
		return state, err
	}
	defer s.Close()

	// Interrogate service
	status, e := s.Control(svc.Interrogate)
	if e != nil {
		if errors.Is(e, windows.ERROR_SERVICE_NOT_ACTIVE) {
			return svc.Stopped, nil
		}
		return state, e
	}

	return status.State, e
}

// StartService starts a service
func StartService(serviceName string) error {
	state, e := GetServiceState(serviceName)

	if e != nil {
		return e
	}

	if state == svc.Running {
		return nil
	}

	// open service manager, requires admin
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	// open the service
	s, err := m.OpenService(serviceName)
	if err != nil {
		return err
	}
	defer s.Close()

	// start the service
	err = s.Start()
	if err != nil {
		return err
	}

	// Services do not immediately start.
	var waitForStatusUpdate int = 5
	var status svc.Status

	for i := 0; i < waitForStatusUpdate; i++ {
		time.Sleep(1 * time.Second)
		status, err := GetServiceState(serviceName)

		if err == nil && status == svc.Running {
			// Returns nil as service is running
			return nil
		}
	}

	return fmt.Errorf("'%s' did not start in time; status: %+v", serviceName, status)
}

// StopService stops a service
func StopService(serviceName string) error {
	state, e := GetServiceState(serviceName)

	if e != nil {
		return e
	}

	if state == svc.Stopped {
		return nil
	}

	// open service manager, requires admin
	m, err := mgr.Connect()
	if nil != err {
		return err
	}
	defer m.Disconnect()

	// open the service
	s, err := m.OpenService(serviceName)
	if err != nil {
		return err
	}
	defer s.Close()

	// stop the service
	_, err = s.Control(svc.Stop)
	if err != nil {
		return err
	}

	// services may request up to 125 seconds of time before being killed
	// https://docs.microsoft.com/en-us/windows/win32/api/winsvc/nc-winsvc-lphandler_function#remarks
	var retries int = 130
	var status svc.Status

	for i := 0; i < retries; i++ {
		time.Sleep(1 * time.Second)
		status, err := GetServiceState(serviceName)

		if err == nil && status == svc.Stopped {
			// Returns nil as service is no longer running
			return nil
		}
	}

	return fmt.Errorf("'%s' did not stop in time; status: %+v", serviceName, status)
}
