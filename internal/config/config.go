package config

import (
	"log"
	"os"
	"strings"
)

type PaycorConfig struct {
	// Paycor Configuration
	PaycorClientID               string
	PaycorClientSecret           string
	PaycorOcpApimSubscriptionKey string // Updated field name
	PaycorRefreshToken           string
	PaycorTokenURLBase           string
	PaycorAPIBaseURL             string
	PaycorLegalEntityID          string
	PaycorScopes                 []string
}

type JiraConfig struct {
	// Jira Configuration
	JiraAssetsURL              string // Base URL for Jira (e.g., https://your-domain.atlassian.net)
	JiraAdminEmail             string
	JiraOrgAPIKey              string
	JiraSiteName               string // e.g., your-company.atlassian.net (used for workspace ID discovery & standard API calls)
	JiraWorkspaceID            string // Assets workspace ID (can be discovered or set via env)
	JiraObjectSchemaKey        string // "HRITBETA"
	JiraEmployeeObjectTypeName string // Name of the Employee Object Type in Assets, e.g., "Employee"
	JiraEmployeeObjectTypeID   string // Discovered or set via env for "Employee" type

	// Jira Issue Creation & Linking Configuration
	JiraTestProjectKey            string // Project key for creating linked Jira issues (e.g., "TEST")
	JiraIssueTypeNameForAsset     string // Name of the issue type to create (e.g., "Task", "Story")
	JiraIssueTypeIDForAsset       string // Discovered or set via env
	JiraLinkTypeNameToAsset       string // Name of the issue link type (e.g., "Relates to", "Impacts")
	JiraLinkTypeIDToAsset         string // Discovered or set via env
	JiraAssetObjectKeyCustomField string // Custom field ID for storing Asset Object Key on Jira issue (e.g. "customfield_10050")
}

// --- Configuration Struct (Combined for Paycor and Jira) ---
type AppConfig struct {
	// Paycor Configuration
	Paycor PaycorConfig // Embedded PaycorConfig struct for modularity
	Jira   JiraConfig   // Embedded JiraConfig struct for modularity
	// General
	LogFilePath string
}

// JiraConfig (placeholder for modularity)
// type JiraConfig struct {
// 	BaseURL  string
// 	APIToken string
// 	// ... other Jira specific configs
// }

// Load loads configuration from environment variables.
// For this focused task, it primarily loads Paycor config.
func Load() (*AppConfig, error) {
	// Read the PAYCOR_SCOPES environment variable and split it into a slice.
	scopesString := getEnv("PAYCOR_SCOPES", "") // Read the new variable
	var scopes []string
	if scopesString != "" {
		scopes = strings.Split(scopesString, ",")
	}
	cfg := &AppConfig{
		Paycor: PaycorConfig{
			PaycorClientID:               getEnv("PAYCOR_CLIENT_ID", ""),
			PaycorClientSecret:           getEnv("PAYCOR_CLIENT_SECRET", ""),
			PaycorOcpApimSubscriptionKey: getEnv("PAYCOR_OCP_APIM_SUBSCRIPTION_KEY", ""),
			PaycorRefreshToken:           getEnv("PAYCOR_REFRESH_TOKEN", ""),
			PaycorTokenURLBase:           getEnv("PAYCOR_TOKEN_URL_BASE", ""),
			PaycorAPIBaseURL:             getEnv("PAYCOR_API_BASE_URL", ""),
			PaycorLegalEntityID:          getEnv("PAYCOR_LEGAL_ENTITY_ID", ""),
			PaycorScopes:                 scopes, // Use the split scopes
		},

		Jira: JiraConfig{
			JiraSiteName:                  getEnv("JIRA_ORG_DOMAIN", ""),
			JiraWorkspaceID:               getEnv("JIRA_WORKSPACE_ID", ""),
			JiraAdminEmail:                getEnv("JIRA_ADMIN_EMAIL", ""),
			JiraOrgAPIKey:                 getEnv("JIRA_ORG_API_KEY", ""),
			JiraAssetsURL:                 getEnv("JIRA_ASSETS_URL", ""),
			JiraObjectSchemaKey:           getEnv("JIRA_OBJECT_SCHEMA_KEY", ""),
			JiraAssetObjectKeyCustomField: getEnv("JIRA_ASSET_OBJECT_KEY_CUSTOM_FIELD_ID", ""),
		},
		// Initialize other AppConfig fields
		// DatabaseURL: getEnv("DATABASE_URL", ""),
		// ServerPort:  getEnv("SERVER_PORT", "8080"), // Default port
	}
	// Validate Paycor configuration
	if cfg.Paycor.PaycorClientID == "" {
		log.Println("CONFIG WARNING: PAYCOR_CLIENT_ID environment variable is not set.")
	}
	if cfg.Paycor.PaycorClientSecret == "" {
		log.Println("CONFIG WARNING: PAYCOR_CLIENT_SECRET environment variable is not set.")
	}
	if cfg.Paycor.PaycorOcpApimSubscriptionKey == "" {
		log.Println("CONFIG WARNING: PAYCOR_SUBSCRIPTION_KEY environment variable is not set.")
	}
	if cfg.Paycor.PaycorRefreshToken == "" {
		log.Println("CONFIG WARNING: PAYCOR_REFRESH_TOKEN environment variable is not set.")
	}
	if cfg.Paycor.PaycorTokenURLBase == "" {
		log.Println("CONFIG WARNING: PAYCOR_TOKEN_BASE_URL environment variable is not set.")
	}
	if cfg.Paycor.PaycorAPIBaseURL == "" {
		log.Println("CONFIG WARNING: PAYCOR_BASE_URL environment variable is not set.")
	}
	if cfg.Paycor.PaycorLegalEntityID == "" {
		log.Println("CONFIG WARNING: PAYCOR_LEGAL_ENTITY_ID environment variable is not set.")
	}
	if cfg.Jira.JiraSiteName == "" {
		log.Println("CONFIG WARNING: JIRA_ORG_DOMAIN environment variable is not set.")
	}
	if cfg.Jira.JiraWorkspaceID == "" {
		log.Println("CONFIG WARNING: JIRA_WORKSPACE_ID environment variable is not set.")
	}
	if cfg.Jira.JiraAdminEmail == "" {
		log.Println("CONFIG WARNING: JIRA_ADMIN_EMAIL environment variable is not set.")
	}
	if cfg.Jira.JiraOrgAPIKey == "" {
		log.Println("CONFIG WARNING: JIRA_ORG_API_KEY environment variable is not set.")
	}
	if cfg.Jira.JiraAssetsURL == "" {
		log.Println("CONFIG WARNING: JIRA_ASSETS_URL environment variable is not set.")
	}
	if cfg.Jira.JiraObjectSchemaKey == "" {
		log.Println("CONFIG WARNING: JIRA_OBJECT_SCHEMA_KEY environment variable is not set.")
	}
	if cfg.Jira.JiraAssetObjectKeyCustomField == "" {
		log.Println("CONFIG WARNING: JIRA_ASSET_OBJECT_KEY_CUSTOM_FIELD_ID environment variable is not set.")
	}

	return cfg, nil
}

func getEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		if defaultValue == "" {
			// If there's no default and the env var is not set, it might be an issue or intended.
			// Log a general warning for unset variables without defaults if that's desired behavior.
			// log.Printf("CONFIG INFO: Environment variable %s not set, no default provided.", key)
		} else {
			log.Printf("CONFIG INFO: Environment variable %s not set, using default value.", key)
		}
		return defaultValue
	}
	return value
}

// getEnvAsInt and getEnvAsDuration can be added back if other config sections need them.
// For this specific task, they are not immediately required by PaycorConfig.
