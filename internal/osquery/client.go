package osquery

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type OsqueryClient struct {
	binaryPath string
}

type InstalledApp struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewOsqueryClient() *OsqueryClient {
	return &OsqueryClient{
		binaryPath: "osqueryi",
	}
}

type SystemInfoResult struct {
	OSVersion      string
	OSName         string
	OSPlatform     string
	OsqueryVersion string
}

func (c *OsqueryClient) GetSystemInfo() (SystemInfoResult, error) {
	result := SystemInfoResult{}

	osQuery := "SELECT version, name, platform FROM os_version;"
	osResult, err := c.executeQuery(osQuery)
	if err != nil {
		return result, fmt.Errorf("failed to get OS details: %w", err)
	}

	var osData []map[string]interface{}
	if err := json.Unmarshal([]byte(osResult), &osData); err != nil {
		return result, fmt.Errorf("failed to parse OS data: %w", err)
	}

	if len(osData) == 0 {
		return result, fmt.Errorf("no OS data returned")
	}

	result.OSVersion = fmt.Sprintf("%v", osData[0]["version"])
	result.OSName = fmt.Sprintf("%v", osData[0]["name"])
	result.OSPlatform = fmt.Sprintf("%v", osData[0]["platform"])

	osqueryVersionQuery := "SELECT version FROM osquery_info;"
	osqueryVersionResult, err := c.executeQuery(osqueryVersionQuery)
	if err != nil {
		return result, fmt.Errorf("failed to get osquery version: %w", err)
	}

	var osqueryVersionData []map[string]interface{}
	if err := json.Unmarshal([]byte(osqueryVersionResult), &osqueryVersionData); err != nil {
		return result, fmt.Errorf("failed to parse osquery version data: %w", err)
	}

	if len(osqueryVersionData) == 0 {
		return result, fmt.Errorf("no osquery version data returned")
	}

	result.OsqueryVersion = fmt.Sprintf("%v", osqueryVersionData[0]["version"])

	return result, nil
}

func (c *OsqueryClient) GetInstalledApps() ([]InstalledApp, error) {
	var query string

	switch runtime.GOOS {
	case "darwin":
		query = "SELECT name, bundle_version AS version FROM apps LIMIT 100;"
	case "windows":
		query = "SELECT name, version FROM programs LIMIT 100;"
	case "linux":
		query = "SELECT name, version FROM deb_packages LIMIT 100;" +
			" UNION SELECT name, version FROM rpm_packages LIMIT 100;"
	default:
		query = "SELECT DISTINCT name, path as version FROM processes LIMIT 100;"
	}

	result, err := c.executeQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get installed apps: %w", err)
	}

	var appsData []map[string]interface{}
	if err := json.Unmarshal([]byte(result), &appsData); err != nil {
		return nil, fmt.Errorf("failed to parse installed apps data: %w", err)
	}

	apps := make([]InstalledApp, 0, len(appsData))
	for _, app := range appsData {
		name, nameOk := app["name"].(string)
		version, versionOk := app["version"].(string)

		if !nameOk {
			continue
		}

		if !versionOk {
			version = "unknown"
		}

		apps = append(apps, InstalledApp{
			Name:    name,
			Version: version,
		})
	}

	return apps, nil
}

func (c *OsqueryClient) executeQuery(query string) (string, error) {
	cmd := exec.Command(c.binaryPath, "--json", query)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("osquery execution failed: %w, output: %s", err, string(output))
	}

	result := strings.TrimSpace(string(output))
	return result, nil
}
