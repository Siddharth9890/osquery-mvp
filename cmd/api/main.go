package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/Siddharth9890/osquery-mvp/config"
	"github.com/Siddharth9890/osquery-mvp/internal/database"
	api "github.com/Siddharth9890/osquery-mvp/internal/handler"
	"github.com/Siddharth9890/osquery-mvp/internal/osquery"
	"github.com/Siddharth9890/osquery-mvp/pkg/logger"
	"github.com/Siddharth9890/osquery-mvp/pkg/middleware"
	"github.com/Siddharth9890/osquery-mvp/ui"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logger.InitLogger(logLevel)
	defer logger.Close()

	log := logger.Log

	log.Info("Starting osquery MVP service",
		zap.String("log_level", logLevel))

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration",
			zap.Error(err))
	}

	log.Info("Configuration loaded",
		zap.String("db_name", cfg.DBName),
		zap.String("api_port", cfg.APIPort),
		zap.Duration("refresh_interval", cfg.RefreshInterval))

	if err := osquery.CheckOsqueryInstallation(); err != nil {
		log.Fatal("Osquery check failed",
			zap.Error(err))
	}
	log.Info("Osquery installation verified successfully")

	log.Debug("Connecting to database...")
	dbConn, err := config.NewDatabaseConnection(cfg.GetDBConnectionString())
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbConn.Close()

	dbService := database.NewService(dbConn)

	querier := osquery.NewOsqueryClient()

	log.Info("Running initial data collection...")
	if err := collectAndStoreData(querier, dbService); err != nil {
		log.Error("Error in initial data collection",
			zap.Error(err))
	}

	requestIDMiddleware := middleware.RequestIDMiddleware

	apiHandler := api.NewHandler(dbService)
	http.Handle("/api/latest_data", requestIDMiddleware(http.HandlerFunc(apiHandler.GetLatestData)))

	uiHandler, err := ui.NewHandler(dbService, "http://localhost:"+cfg.APIPort+"/api")
	if err != nil {
		log.Fatal("Failed to create UI handler",
			zap.Error(err))
	}
	http.Handle("/", requestIDMiddleware(http.HandlerFunc(uiHandler.Dashboard)))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("ui/assets"))))

	go func() {
		log.Info("Starting server",
			zap.String("address", cfg.GetAPIAddress()))
		log.Info("Endpoints available",
			zap.String("api_endpoint", "http://localhost:"+cfg.APIPort+"/api/latest_data"),
			zap.String("ui_dashboard", "http://localhost:"+cfg.APIPort+"/"))

		if err := http.ListenAndServe(cfg.GetAPIAddress(), nil); err != nil {
			log.Fatal("Failed to start server",
				zap.Error(err))
		}
	}()

	ticker := time.NewTicker(cfg.RefreshInterval)
	defer ticker.Stop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			log.Info("Running scheduled data collection...")
			if err := collectAndStoreData(querier, dbService); err != nil {
				log.Error("Error in scheduled data collection",
					zap.Error(err))
			}
		case <-stop:
			log.Info("Shutting down...")
			return
		}
	}
}

func collectAndStoreData(querier *osquery.OsqueryClient, dbService *database.Service) error {
	log := logger.Log

	log.Debug("Querying system information from osquery")
	sysInfo, err := querier.GetSystemInfo()
	if err != nil {
		log.Error("Failed to get system information from osquery",
			zap.Error(err))
		return err
	}

	log.Debug("Querying installed applications from osquery")
	apps, err := querier.GetInstalledApps()
	if err != nil {
		log.Error("Failed to get installed applications from osquery",
			zap.Error(err))
		return err
	}

	log.Debug("Storing collected data in database",
		zap.Int("app_count", len(apps)))
	if err := dbService.StoreSystemInfo(sysInfo, apps); err != nil {
		log.Error("Failed to store data in database",
			zap.Error(err))
		return err
	}

	log.Info("Data collection completed successfully",
		zap.String("os_version", sysInfo.OSVersion),
		zap.String("os_name", sysInfo.OSName),
		zap.String("os_platform", sysInfo.OSPlatform),
		zap.String("osquery_version", sysInfo.OsqueryVersion),
		zap.Int("app_count", len(apps)))
	return nil
}
