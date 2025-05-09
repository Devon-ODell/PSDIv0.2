package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Version  string
	Server   ServerConfig
	Database DatabaseConfig
	Jira     JiraConfig
	Paycor   PaycorConfig
	Worker   WorkerConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

// JiraConfig holds Jira API configuration
type JiraConfig struct {
	BaseURL     string
	APIToken    string
	UserEmail   string
	AssetObject string
}

// PaycorConfig holds Paycor API configuration
type PaycorConfig struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
}

// WorkerConfig holds background worker configuration
type WorkerConfig struct {
	IntervalSeconds int
	MaxRetries      int
	RetryDelay      time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Set defaults
	config := &Config{
		Version: getEnv("APP_VERSION", "1.0.0"),
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", "15s"),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", "15s"),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", "30s"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "paycorjira"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Jira: JiraConfig{
			BaseURL:     getEnv("JIRA_BASE_URL", ""),
			APIToken:    getEnv("JIRA_API_TOKEN", ""),
			UserEmail:   getEnv("JIRA_USER_EMAIL", ""),
			AssetObject: getEnv("JIRA_ASSET_OBJECT", "Employee"),
		},
		Paycor: PaycorConfig{
			ClientID:     getEnv("PAYCOR_CLIENT_ID", ""),
			ClientSecret: getEnv("PAYCOR_CLIENT_SECRET", ""),
			BaseURL:      getEnv("PAYCOR_BASE_URL", "https://apis.paycor.com"),
		},
		Worker: WorkerConfig{
			IntervalSeconds: getEnvAsInt("WORKER_INTERVAL_SECONDS", "30"),
			MaxRetries:      getEnvAsInt("WORKER_MAX_RETRIES", "5"),
			RetryDelay:      getEnvAsDuration("WORKER_RETRY_DELAY", "5m"),
		},
	}

	// Validate required configuration
	if config.Jira.BaseURL == "" {
		return nil, fmt.Errorf("JIRA_BASE_URL is required")
	}
	if config.Jira.APIToken == "" {
		return nil, fmt.Errorf("JIRA_API_TOKEN is required")
	}
	if config.Jira.UserEmail == "" {
		return nil, fmt.Errorf("JIRA_USER_EMAIL is required")
	}
	if config.Paycor.ClientID == "" {
		return nil, fmt.Errorf("PAYCOR_CLIENT_ID is required")
	}
	if config.Paycor.ClientSecret == "" {
		return nil, fmt.Errorf("PAYCOR_CLIENT_SECRET is required")
	}

	return config, nil
}

// Helper functions for environment variables
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key, fallback string) int {
	valueStr := getEnv(key, fallback)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		value, _ = strconv.Atoi(fallback)
	}
	return value
}

func getEnvAsDuration(key, fallback string) time.Duration {
	valueStr := getEnv(key, fallback)
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		value, _ = time.ParseDuration(fallback)
	}
	return value
}
