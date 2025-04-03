package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string

	APIPort         string
	RefreshInterval time.Duration
}

func LoadConfig() (*Config, error) {
	godotenv.Load()

	config := &Config{
		DBUser:     getEnv("DB_USER", "osquery"),
		DBPassword: getEnv("DB_PASSWORD", "osquery"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "osquery_data"),

		APIPort: getEnv("API_PORT", "8080"),
	}

	refreshStr := getEnv("REFRESH_INTERVAL", "15m")
	refreshInterval, err := time.ParseDuration(refreshStr)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh interval format: %v", err)
	}
	config.RefreshInterval = refreshInterval

	return config, nil
}

func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func (c *Config) GetAPIAddress() string {
	return ":" + c.APIPort
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
