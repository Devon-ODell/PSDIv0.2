package main

import (
	"context"
	"encoding/json"

	//"fmt"
	"log"
	"os"
	"time"

	// Use your project's actual module path for internal packages
	"github.com/Devon-ODell/PSDIv0.2/internal/config"
	"github.com/Devon-ODell/PSDIv0.2/internal/paycor"
)

func main() {
	// Setup logger
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("INFO: Starting Paycor data extraction process...")

	// Load configuration
	cfg, err := config.Load() // This now primarily loads cfg.Paycor
	if err != nil {
		log.Fatalf("FATAL: Failed to load configuration: %v", err)
	}
	log.Println("INFO: Configuration loaded successfully.")
	log.Printf("DEBUG: Paycor Config: ClientID=%s..., APIBaseURL=%s, LegalEntityID=%s, OutputFile=%s",
		safeSubstring(cfg.Paycor.PaycorClientID, 5), cfg.Paycor.PaycorAPIBaseURL, cfg.Paycor.PaycorLegalEntityID, cfg.LogFilePath)

	// Initialize Paycor client
	// Pass the Paycor-specific config struct to the client
	paycorClient, err := paycor.NewClient(context.Background(), cfg.Paycor)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize Paycor client: %v", err)
	}
	log.Println("INFO: Paycor client initialized successfully.")

	// --- Step 1: Fetch Employees from Paycor ---
	log.Println("INFO: Attempting to fetch all employees from Paycor...")
	startTime := time.Now()
	employees, err := paycorClient.FetchAllEmployees(context.Background())
	if err != nil {
		log.Fatalf("FATAL: Failed to fetch employees from Paycor: %v", err)
	}
	duration := time.Since(startTime)
	log.Printf("INFO: Successfully fetched %d employees from Paycor in %v.", len(employees), duration)

	// --- Step 2: Marshal employee data to JSON ---
	log.Println("INFO: Marshalling employee data to JSON...")
	jsonData, err := json.MarshalIndent(employees, "", "  ") // Pretty print JSON
	if err != nil {
		log.Fatalf("FATAL: Failed to marshal employee data to JSON: %v", err)
	}
	log.Printf("INFO: Employee data successfully marshalled to JSON (%d bytes).", len(jsonData))

	// --- Step 3: Save JSON data to a file ---
	outputFilePath := cfg.LogFilePath
	log.Printf("INFO: Attempting to save JSON data to file: %s", outputFilePath)
	err = os.WriteFile(outputFilePath, jsonData, 0644) // rw-r--r-- permissions
	if err != nil {
		log.Fatalf("FATAL: Failed to write JSON data to file '%s': %v", outputFilePath, err)
	}
	log.Printf("SUCCESS: Employee data successfully saved to %s", outputFilePath)

	// --- (Detached) Placeholder for Jira Integration ---
	// This section is for conceptual planning and will not be executed in this flow.
	/*
	   log.Println("INFO: (Placeholder) Beginning Jira integration phase...")

	   // 1. Initialize Jira Client (would require JiraConfig)
	   // jiraCfg := cfg.Jira // Assuming JiraConfig is part of the main Config struct
	   // jiraClient, err := jira.NewClient(jiraCfg)
	   // if err != nil {
	   //     log.Fatalf("FATAL: Failed to initialize Jira client: %v", err)
	   // }
	   // log.Println("INFO: (Placeholder) Jira client initialized.")

	   // 2. Process each Paycor employee and interact with Jira
	   // for _, emp := range employees {
	   //     log.Printf("INFO: (Placeholder) Processing employee ID %s for Jira sync...", emp.ID)
	   //
	   //     // Example: Map Paycor employee data to Jira issue fields
	   //     // jiraIssueData := mapPaycorToJira(emp)
	   //
	   //     // Example: Check if asset exists in Jira, then create or update
	   //     // assetExists, err := jiraClient.CheckAssetExists(emp.ID) // Hypothetical method
	   //     // if err != nil {
	   //     //     log.Printf("ERROR: (Placeholder) Failed to check Jira asset for employee %s: %v", emp.ID, err)
	   //     //     continue
	   //     // }
	   //
	   //     // if assetExists {
	   //     //     err = jiraClient.UpdateAsset(emp.ID, jiraIssueData) // Hypothetical method
	   //     // } else {
	   //     //     err = jiraClient.CreateAsset(jiraIssueData) // Hypothetical method
	   //     // }
	   //     // if err != nil {
	   //     //     log.Printf("ERROR: (Placeholder) Failed to sync employee %s to Jira: %v", emp.ID, err)
	   //     // } else {
	   //     //     log.Printf("INFO: (Placeholder) Successfully synced employee %s to Jira.", emp.ID)
	   //     // }
	   // }
	   // log.Println("INFO: (Placeholder) Jira integration phase completed.")
	*/

	log.Println("INFO: Paycor data extraction process finished successfully. Exiting.")
}

func safeSubstring(s string, length int) string {
	if len(s) < length {
		return s
	}
	return s[:length]
}

// Placeholder for mapping function if you were to integrate with Jira
// func mapPaycorToJira(employee paycor.Employee) map[string]interface{} {
//     // Implementation would depend on your Jira setup and desired field mappings
//     return map[string]interface{}{
//         "summary": fmt.Sprintf("%s %s (%s)", employee.FirstName, employee.LastName, employee.ID),
//         "customfield_XXXXX": employee.WorkEmail, // Example custom field
//         // ... other fields
//     }
// }
