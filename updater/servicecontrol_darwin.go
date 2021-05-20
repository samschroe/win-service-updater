// +build darwin

// Dummy file so I could develop on a Mac

package updater

// IsServiceRunning checks to see if a service is running
func IsServiceRunning(serviceName string) (bool, error) {
	println("calling IsServiceRunning stub with", serviceName)
	return true, nil
}

// StartService starts a service
func StartService(serviceName string) error {
	println("calling StartService stub with", serviceName)
	return nil
}

// StopService stops a service
func StopService(serviceName string) error {
	println("calling StopService stub with", serviceName)
	return nil
}
