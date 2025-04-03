package database

import (
	"database/sql"
	"fmt"
	"log"

	model "github.com/Siddharth9890/osquery-mvp/internal/models"
	"github.com/Siddharth9890/osquery-mvp/internal/osquery"
	"github.com/Siddharth9890/osquery-mvp/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Close() error {
	return s.db.Close()
}

func (s *Service) StoreSystemInfo(sysInfo osquery.SystemInfoResult, apps []osquery.InstalledApp) error {
	log := logger.Log.With(
		zap.String("os_version", sysInfo.OSVersion),
		zap.String("osquery_version", sysInfo.OsqueryVersion),
	)

	log.Debug("Starting database transaction for system info storage")

	tx, err := s.db.Begin()
	if err != nil {
		log.Error("Failed to begin database transaction",
			zap.Error(err))
		return fmt.Errorf("transaction error: %w", err)
	}

	defer func() {
		if err != nil {
			log.Debug("Rolling back transaction due to error")
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Error("Failed to rollback transaction",
					zap.Error(rbErr))
			}
		}
	}()

	log.Debug("Inserting system info record")
	result, err := tx.Exec(
		"INSERT INTO system_info (os_version, os_name, os_platform, osquery_version) VALUES (?, ?, ?, ?)",
		sysInfo.OSVersion, sysInfo.OSName, sysInfo.OSPlatform, sysInfo.OsqueryVersion,
	)
	if err != nil {
		log.Error("Failed to insert system info record",
			zap.Error(err))
		return fmt.Errorf("database insert error: %w", err)
	}

	systemInfoID, err := result.LastInsertId()
	if err != nil {
		log.Error("Failed to get last insert ID",
			zap.Error(err))
		return fmt.Errorf("database ID retrieval error: %w", err)
	}

	log.Debug("System info record created",
		zap.Int64("system_info_id", systemInfoID))

	log.Debug("Inserting installed apps records")
	for i, app := range apps {
		_, err := tx.Exec(
			"INSERT INTO installed_apps (system_info_id, name, version) VALUES (?, ?, ?)",
			systemInfoID, app.Name, app.Version,
		)
		if err != nil {
			log.Error("Failed to insert app record",
				zap.Error(err),
				zap.String("app_name", app.Name),
				zap.Int("app_index", i))
			return fmt.Errorf("database insert error for app '%s': %w", app.Name, err)
		}
	}

	log.Debug("Committing transaction")
	if err := tx.Commit(); err != nil {
		log.Error("Failed to commit transaction",
			zap.Error(err))
		return fmt.Errorf("transaction commit error: %w", err)
	}

	log.Info("Successfully stored system info and apps in database",
		zap.Int64("system_info_id", systemInfoID))
	return nil
}

func (s *Service) GetLatestSystemInfo() (*model.SystemInfo, error) {
	var info model.SystemInfo
	err := s.db.QueryRow(`
		SELECT id, os_version, os_name, os_platform, osquery_version, collected_at 
		FROM system_info 
		ORDER BY collected_at DESC 
		LIMIT 1
	`).Scan(&info.ID, &info.OSVersion, &info.OSName, &info.OSPlatform, &info.OsqueryVersion, &info.CollectedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest system info: %w", err)
	}

	rows, err := s.db.Query(`
		SELECT name, version 
		FROM installed_apps 
		WHERE system_info_id = ?
	`, info.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get installed apps: %w", err)
	}
	defer rows.Close()

	info.Apps = []osquery.InstalledApp{}
	for rows.Next() {
		var app osquery.InstalledApp
		if err := rows.Scan(&app.Name, &app.Version); err != nil {
			return nil, fmt.Errorf("failed to scan app row: %w", err)
		}
		info.Apps = append(info.Apps, app)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over app rows: %w", err)
	}

	return &info, nil
}

func (s *Service) GetOSDetails() (string, string, error) {
	osName := "Unknown"
	osPlatform := "Unknown"

	err := s.db.QueryRow(`
		SELECT os_name FROM system_info
		ORDER BY collected_at DESC
		LIMIT 1
	`).Scan(&osName)
	if err != nil {
		log.Printf("Error fetching OS name: %v", err)
	}

	err = s.db.QueryRow(`
		SELECT os_platform FROM system_info
		ORDER BY collected_at DESC
		LIMIT 1
	`).Scan(&osPlatform)
	if err != nil {
		log.Printf("Error fetching OS platform: %v", err)
	}

	return osName, osPlatform, nil
}
