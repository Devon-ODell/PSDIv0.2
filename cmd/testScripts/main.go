// cmd/testscript/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Devon-ODell/PSDIv0.2/internal/config"
	"github.com/Devon-ODell/PSDIv0.2/internal/jira"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file.
	err := godotenv.Load()
	if err != nil {
		log.Println("INFO: No .env file found, relying on OS environment variables.")
	}

	// Setup logger
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("INFO: Starting Jira Role and Issue creation test script...")

	// --- 1. Configuration Loading ---
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("FATAL: Failed to load configuration: %v", err)
	}
	log.Println("INFO: Configuration loaded successfully.")

	// --- 2. Initialize Jira Client ---
	jiraClient, err := jira.NewClient(cfg.Jira)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize Jira client: %v", err)
	}
	log.Println("INFO: Jira client initialized successfully.")

	ctx := context.Background()

	// --- 3. Define and Create Role Asset ---
	// Use a unique name to ensure it's a new role each time the script runs.
	newRoleName := fmt.Sprintf("Test Role %d", time.Now().Unix())
	log.Printf("INFO: Attempting to create a new Role asset named: '%s'", newRoleName)

	roleObjectKey, err := jiraClient.FindOrCreateRole(ctx, newRoleName)
	if err != nil {
		log.Fatalf("FATAL: Failed to create role asset '%s': %v", newRoleName, err)
	}
	if roleObjectKey == "" {
		log.Fatalf("FATAL: Role asset creation for '%s' returned an empty object key.", newRoleName)
	}
	log.Printf("SUCCESS: Successfully created/found Role asset with Key: %s", roleObjectKey)

	// --- 4. Create Jira Issue Referencing the Role ---
	log.Printf("INFO: Attempting to create a Jira issue in project '%s' referencing Role asset '%s'", cfg.Jira.JiraTestProjectKey, roleObjectKey)

	issueSummary := fmt.Sprintf("New Role Provisioned: %s", newRoleName)
	issueDescription := fmt.Sprintf("This is an automated ticket to track the provisioning of the new role asset: %s.", roleObjectKey)

	// This is the custom field ID for your "Asset" custom field.
	// Ensure this is set correctly in your .env file and config.
	assetCustomFieldID := cfg.Jira.JiraAssetObjectKeyCustomField
	if assetCustomFieldID == "" {
		log.Fatal("FATAL: JIRA_ASSET_OBJECT_KEY_CUSTOM_FIELD_ID is not set in the configuration.")
	}

	issueResponse, err := jiraClient.CreateIssueWithAsset(ctx, cfg.Jira.JiraTestProjectKey, issueSummary, issueDescription, assetCustomFieldID, roleObjectKey)
	if err != nil {
		log.Fatalf("FATAL: Failed to create Jira issue: %v", err)
	}

	log.Printf("SUCCESS: Successfully created Jira issue with Key: %s", issueResponse.Key)
	log.Printf("INFO: View the new issue here: https://%s/browse/%s", cfg.Jira.JiraSiteName, issueResponse.Key)
	log.Println("INFO: Test script finished successfully.")
}
