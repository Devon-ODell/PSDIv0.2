package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv" // Added for .env loading convenience
)

// Config holds all configuration for the application
type Config struct {
	Paycor PaycorConfig
	// Add other configs like JiraConfig if needed for future modular parts
	// For now, we only strictly need PaycorConfig for the current task.
	// Jira JiraConfig
}

// PaycorConfig holds Paycor API configuration
type PaycorConfig struct {
	ClientID        string
	ClientSecret    string
	SubscriptionKey string
	RefreshToken    string
	TokenURLBase    string   // e.g., https://apis-sandbox.paycor.com/sts/v1/token
	APIBaseURL      string   // e.g., https://apis-sandbox.paycor.com/v1
	LegalEntityID   string   // The specific Legal Entity ID to fetch employees for
	Scopes          []string // Optional scopes for OAuth2
	OutputFilePath  string   // Optional: Path to save the JSON output
}

// JiraConfig (placeholder for modularity)
// type JiraConfig struct {
// 	BaseURL  string
// 	APIToken string
// 	// ... other Jira specific configs
// }

// Load loads configuration from environment variables.
// For this focused task, it primarily loads Paycor config.
func Load() (*Config, error) {
	// Attempt to load .env file from the current directory or parent directories.
	// Useful for local development.
	err := godotenv.Load()
	if err != nil {
		log.Printf("INFO: [Config] No .env file found or error loading it: %v. Relying on OS environment variables if set.", err)
	}

	paycorCfg := PaycorConfig{
		ClientID:        getEnv("PAYCOR_CLIENT_ID", ""),
		ClientSecret:    getEnv("PAYCOR_CLIENT_SECRET", ""),
		SubscriptionKey: getEnv("PAYCOR_SUBSCRIPTION_KEY", ""),
		RefreshToken:    getEnv("PAYCOR_REFRESH_TOKEN", ""),
		TokenURLBase:    getEnv("PAYCOR_TOKEN_BASE_URL", ""),
		APIBaseURL:      getEnv("PAYCOR_BASE_URL", ""),
		LegalEntityID:   getEnv("PAYCOR_LEGAL_ENTITY_ID", ""),
		OutputFilePath:  getEnv("PAYCOR_OUTPUT_FILE_PATH", "paycor_employees.json"), // Default output file path
	}

	// Validate Paycor configuration
	var missingVars []string
	if paycorCfg.ClientID == "" {
		missingVars = append(missingVars, "PAYCOR_CLIENT_ID")
	}
	if paycorCfg.ClientSecret == "" {
		missingVars = append(missingVars, "PAYCOR_CLIENT_SECRET")
	}
	if paycorCfg.SubscriptionKey == "" {
		missingVars = append(missingVars, "PAYCOR_SUBSCRIPTION_KEY")
	}
	if paycorCfg.RefreshToken == "" {
		missingVars = append(missingVars, "PAYCOR_REFRESH_TOKEN")
	}
	if paycorCfg.TokenURLBase == "" {
		missingVars = append(missingVars, "PAYCOR_TOKEN_URL_BASE")
	}
	if paycorCfg.APIBaseURL == "" {
		missingVars = append(missingVars, "PAYCOR_API_BASE_URL")
	}
	if paycorCfg.LegalEntityID == "" {
		missingVars = append(missingVars, "PAYCOR_LEGAL_ENTITY_ID")
	}

	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing required Paycor environment variable(s): %s", strings.Join(missingVars, ", "))
	}

	return &Config{Paycor: paycorCfg}, nil
}

// Helper functions for environment variables
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvAsInt and getEnvAsDuration can be added back if other config sections need them.
// For this specific task, they are not immediately required by PaycorConfig.
