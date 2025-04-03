package models

import (
	"time"

	"github.com/Siddharth9890/osquery-mvp/internal/osquery"
)

type SystemInfo struct {
	ID             int                    `json:"id"`
	OSVersion      string                 `json:"os_version"`
	OSName         string                 `json:"os_name"`
	OSPlatform     string                 `json:"os_platform"`
	OsqueryVersion string                 `json:"osquery_version"`
	CollectedAt    time.Time              `json:"collected_at"`
	Apps           []osquery.InstalledApp `json:"installed_apps"`
}
