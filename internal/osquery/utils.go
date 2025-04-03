package osquery

import (
	"fmt"
	"os/exec"
	"strings"
)

func CheckOsqueryInstallation() error {
	cmd := exec.Command("osqueryi", "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("osquery is not installed or not in PATH. Please install osquery: https://osquery.io/downloads")
		}

		return fmt.Errorf("error checking osquery installation: %w (output: %s)", err, string(output))
	}

	if !strings.Contains(string(output), "osqueryi version") {
		return fmt.Errorf("osquery appears to be installed but returned unexpected output: %s", string(output))
	}

	return nil
}
