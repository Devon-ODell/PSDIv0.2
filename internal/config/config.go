package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv" // Added for .env loading convenience
)

// --- Configuration Struct (Combined for Paycor and Jira) ---
type AppConfig struct {
	// Paycor Configuration
	PaycorClientID               string
	PaycorClientSecret           string
	PaycorOcpApimSubscriptionKey string // Updated field name
	PaycorRefreshToken           string
	PaycorTokenURLBase           string
	PaycorAPIBaseURL             string
	PaycorLegalEntityID          string
	PaycorScopes                 []string

	// Jira Configuration
	JiraAssetsURL                    string // Base URL for Jira (e.g., https://your-domain.atlassian.net)
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

	// Jira Attribute IDs for Employee Object (CRITICAL - map these in .env)
	JiraAttrIDPaycorSystemID string
	JiraAttrIDEmployeeID     string
	JiraAttrIDFirstName      string
	JiraAttrIDLastName       string
	JiraAttrIDEmail          string
	JiraAttrIDDepartment     string
	JiraAttrIDJobTitle       string
	JiraAttrIDLocation       string
	JiraAttrIDStartDate      string
	// Add other attribute IDs as needed

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
	// Attempt to load .env file from the current directory or parent directories.
	// Useful for local development.
	err := godotenv.Load()
	if err != nil {
		log.Printf("INFO: [Config] No .env file found or error loading it: %v. Relying on OS environment variables if set.", err)
	}

	paycorCfg := AppConfig{
		PaycorClientID:        getEnv("PAYCOR_CLIENT_ID", ""),
		PaycorClientSecret:    getEnv("PAYCOR_CLIENT_SECRET", ""),
		PaycorOcpApimSubscriptionKey: getEnv("PAYCOR_SUBSCRIPTION_KEY", ""),
		PaycorRefreshToken:    getEnv("PAYCOR_REFRESH_TOKEN", ""),
		PaycorTokenURLBase:    getEnv("PAYCOR_TOKEN_BASE_URL", ""),
		PaycorAPIBaseURL:      getEnv("PAYCOR_BASE_URL", ""),
		PaycorLegalEntityID:   getEnv("PAYCOR_LEGAL_ENTITY_ID", ""),
		LogFilePath:  getEnv("LOG_FILE_PATH", "paycor_employees.json"), // Default output file path
	}

	jiraCfg := AppConfig{
		JiraSiteName:        getEnv("JIRA_ORG_DOMAIN", ""),
		JiraWorkspaceID:    getEnv("JIRA_WORKSPACE_ID", ""),
		JiraAdminEmail: getEnv("JIRA_ADMIN_EMAIL", ""),
		JiraOrgAPIKey:    getEnv("JIRA_ORG_API_KEY", ""),
		JiraAssetsURL:    getEnv("JIRA_ASSETS_URL", ""),
		JiraObjectSchemaKey:      getEnv("JIRA_OBJECT_SCHEMA_KEY", ""),
		JiraAssetObjectKeyCustomField:   getEnv("JIRA_ASSET_OBJECT_KEY_CUSTOM_FIELD_ID", ""),
		LogFilePath:  getEnv("LOG_FILE_PATH", "jira_assets.json"), // Default output file path
	}

	// Validate Paycor configuration
	var missingVars []string
	if paycorCfg.PaycorClientID == "" {
		missingVars = append(missingVars, "PAYCOR_CLIENT_ID")}
	if paycorCfg.PaycorClientSecret == "" {
		missingVars = append(missingVars, "PAYCOR_CLIENT_SECRET")}
	if paycorCfg.PaycorOcpApimSubscriptionKey == "" {
		missingVars = append(missingVars, "PAYCOR_SUBSCRIPTION_KEY")}
	if paycorCfg.PaycorRefreshToken == "" {
		missingVars = append(missingVars, "PAYCOR_REFRESH_TOKEN")}
	if paycorCfg.PaycorTokenURLBase == "" {
		missingVars = append(missingVars, "PAYCOR_TOKEN_URL_BASE")}
	if paycorCfg.PaycorAPIBaseURL == "" {
		missingVars = append(missingVars, "PAYCOR_API_BASE_URL")}
	if paycorCfg.PaycorLegalEntityID == "" {
		missingVars = append(missingVars, "PAYCOR_LEGAL_ENTITY_ID")}
	if jiraCfg.JiraSiteName == "" {
		missingVars = append(missingVars, "JIRA_ORG_DOMAIN")}
	if jiraCfg.JiraWorkspaceID == "" {
		missingVars = append(missingVars, "JIRA_WORKSPACE_ID")}
	if jiraCfg.JiraAdminEmail == "" {
		missingVars = append(missingVars, "JIRA_ADMIN_EMAIL")}
	if jiraCfg.JiraOrgAPIKey == "" {
		missingVars = append(missingVars, "JIRA_ORG_API_KEY")}
	if jiraCfg.JiraAssetsURL == "" {
		missingVars = append(missingVars, "JIRA_ASSETS_URL")}
	if jiraCfg.JiraObjectSchemaKey == "" {
		missingVars = append(missingVars, "JIRA_OBJECT_SCHEMA_KEY")}
	if jiraCfg.JiraAssetObjectKeyCustomField == "" {
		missingVars = append(missingVars, "JIRA_ASSET_OBJECT_KEY_CUSTOM_FIELD_ID")

	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing required Paycor environment variable(s): %s", strings.Join(missingVars, ", "))
	}

	
	return &AppConfig{Paycor: paycorCfg}, nil
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
