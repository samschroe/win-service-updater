package updater

import (
	"os"
	"os/exec"
)

// Stubbed in function so that tests pass on Darwin
func DoesServiceExist(serviceName string) (bool, error) {
	return false, nil
}

// launchctl is a helper method to execute the MacOS launchctl
// executable. That executable is used to control services on MacOS.
func launchctl(cmdAndArgs ...string) error {
	const launchCtl = "/bin/launchctl"

	cmd := exec.Command(launchCtl, cmdAndArgs...)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// amIAdmin return whether or not user running the program is root
func amIAdmin() bool {
	return os.Geteuid() == 0
}

// IsServiceRunning checks to see if a service is running. The
// serviceName argument is a launchctl domain/<service name>
// specifier. For the Huntress agent it will be
// system/com.huntresslabs.agent
func IsServiceRunning(serviceName string) (bool, error) {
	err := launchctl("print", serviceName)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// StartService starts a service. The servicePlist argument is
// typically the name of a .plist file in the /Library/LaunchDaemons
// directory
func StartService(servicePlist string) error {
	err := launchctl("bootstrap", "system", servicePlist)
	if err != nil {
		return err
	}
	return nil
}

// StopService stops a service
func StopService(servicePlist string) error {
	err := launchctl("bootout", "system", servicePlist)
	if err != nil {
		return err
	}
	return nil
}
