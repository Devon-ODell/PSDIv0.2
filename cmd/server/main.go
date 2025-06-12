package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	// Import the new godotenv package
	"github.com/joho/godotenv"

	// Use your project's actual module path for internal packages
	"github.com/Devon-ODell/PSDIv0.2/internal/config"
	"github.com/Devon-ODell/PSDIv0.2/internal/jira"   // <-- IMPORT for Jira client
	"github.com/Devon-ODell/PSDIv0.2/internal/models" // <-- IMPORT for shared data models
	"github.com/Devon-ODell/PSDIv0.2/internal/paycor"
)

func main() {
	// Load .env file. Not fatal if it doesn't exist.
	err := godotenv.Load()
	if err != nil {
		log.Println("INFO: No .env file found, relying on OS environment variables.")
	}

	// Setup logger
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("INFO: Starting Paycor data extraction and Jira sync process...")

	// =========================================================================
	// Configuration Loading
	// =========================================================================
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("FATAL: Failed to load configuration: %v", err)
	}
	log.Println("INFO: Configuration loaded successfully.")

	// Create a background context for our API calls
	ctx := context.Background()

	// =========================================================================
	// Paycor Data Extraction
	// =========================================================================
	// Initialize Paycor client
	paycorClient, err := paycor.NewClient(ctx, cfg.Paycor)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize Paycor client: %v", err)
	}
	log.Println("INFO: Paycor client initialized successfully.")

	// Fetch all employees from Paycor
	log.Println("INFO: Attempting to fetch all employees from Paycor...")
	startTime := time.Now()
	employees, err := paycorClient.FetchAllEmployees(ctx)
	if err != nil {
		log.Fatalf("FATAL: Failed to fetch employees from Paycor: %v", err)
	}
	duration := time.Since(startTime)
	log.Printf("INFO: Successfully fetched %d employees from Paycor in %v.", len(employees), duration)

	// If no employees are found, there's nothing to sync. Exit gracefully.
	if len(employees) == 0 {
		log.Println("INFO: No employees found in Paycor. Nothing to sync to Jira. Exiting.")
		return
	}

	// Optional: Save the fetched data to a local JSON file for debugging
	saveDataToFile("paycor_employees.json", employees)

	// =========================================================================
	// Jira Integration and Syncing
	// =========================================================================
	log.Println("INFO: Beginning Jira integration phase...")

	// 1. Initialize Jira Client using the Jira-specific config
	jiraClient, err := jira.NewClient(cfg.Jira)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize Jira client: %v", err)
	}
	log.Println("INFO: Jira client initialized successfully.")

	// 2. Fetch all existing Employee Assets from Jira
	// This is done once to avoid making a request for every single employee in the loop.
	log.Println("INFO: Fetching all existing employee assets from Jira for comparison...")
	existingJiraAssets, err := jiraClient.GetAllEmployeeAssets(ctx)
	if err != nil {
		log.Fatalf("FATAL: Failed to get employee assets from Jira: %v", err)
	}
	log.Printf("INFO: Found %d existing employee assets in Jira.", len(existingJiraAssets))

	// 3. Create a map for efficient lookups using the employee's email as a unique key.
	jiraAssetsMap := make(map[string]models.EmployeeAssets)
	for _, asset := range existingJiraAssets {
		// This is the correct way to get the email
		if email := findEmailInAttributes(asset.Attributes); email != "" {
			jiraAssetsMap[email] = asset
		}
	}

	// 4. Loop through Paycor employees and sync to Jira
	log.Println("INFO: Starting sync process for each Paycor employee...")
	for _, emp := range employees {
		log.Printf("INFO: Processing Paycor employee: %s %s (Email: %s)", emp.FirstName, emp.LastName, emp.Email.EmailAddress)

		roleKey, err := jiraClient.FindOrCreateRole(ctx, emp.PositionData.JobTitle)
		if err != nil {
			log.Printf("ERROR: Could not find or create Jira Role for '%s'. Skipping this employee. Error: %v", emp.PositionData.JobTitle, err)
			continue // Skip to the next employee
		}
		if roleKey == "" {
			log.Printf("WARN: No role key was found or created for job title '%s'. The 'Job Role' field will be empty.", emp.PositionData.JobTitle)
		}

		// Map Paycor data to the structure Jira expects
		jiraAssetData := mapPaycorToJiraAsset(emp, roleKey)

		// Check if an asset with this email already exists in our map
		existingAsset, exists := jiraAssetsMap[emp.Email.EmailAddress]

		if exists {
			// UPDATE: The asset already exists, so we update it.
			log.Printf("INFO: Employee exists in Jira. Updating asset ID %s.", existingAsset.ID)
			err = jiraClient.UpdateEmployeeAsset(ctx, existingAsset.ID, jiraAssetData)
			if err != nil {
				log.Printf("ERROR: Failed to update Jira asset for employee %s: %v", emp.ID, err)
			} else {
				log.Printf("SUCCESS: Successfully updated Jira asset for employee %s.", emp.ID)
			}
		} else {
			// CREATE: The asset does not exist, so we create a new one.
			log.Println("INFO: Employee does not exist in Jira. Creating new asset.")
			newAssetID, err := jiraClient.CreateEmployeeAsset(ctx, jiraAssetData)
			if err != nil {
				log.Printf("ERROR: Failed to create Jira asset for employee %s: %v", emp.ID, err)
			} else {
				log.Printf("SUCCESS: Successfully created new Jira asset for employee %s with ID %s.", emp.ID, newAssetID)
			}
		}
	}

	log.Println("INFO: Jira integration phase completed.")
	log.Println("INFO: Process finished successfully. Exiting.")
}

// mapPaycorToJiraAsset converts a Paycor employee object to the Jira EmployeeAssets model.
// !!! IMPORTANT !!!
// You MUST customize the map keys (e.g., "Name", "First Name", "Last Name") to match
//
// mapPaycorToJiraAsset converts a Paycor employee object to the Jira EmployeeAssets model.
// This function now builds the correct []AssetAttribute slice structure.
func mapPaycorToJiraAsset(employee models.Employee, roleKey string) models.EmployeeAssets {
	// !!! IMPORTANT !!!
	// The 'ObjectTypeAttributeID' values below (e.g., "1086", "1093") are based on the
	// 'jiraAssetMap.go' file you provided. You MUST verify these IDs are correct
	// for your specific Jira Assets schema. You can find them in the Jira UI
	// when configuring your object schema.
	return models.EmployeeAssets{
		Attributes: []models.AssetAttribute{
			{
				ObjectTypeAttributeID: strconv.Itoa(models.AttributeID["Name"]), // "1086"
				Values: []models.Value{
					{Value: fmt.Sprintf("%s %s", employee.FirstName, employee.LastName)},
				},
			},
			{
				ObjectTypeAttributeID: strconv.Itoa(models.AttributeID["Email"]), // "1093"
				Values: []models.Value{
					{Value: employee.Email.EmailAddress},
				},
			},
			{
				ObjectTypeAttributeID: strconv.Itoa(models.AttributeID["Start Date"]), // "1095"
				Values: []models.Value{
					{Value: employee.EmploymentDateData.HireDate},
				},
			},
			{
				ObjectTypeAttributeID: strconv.Itoa(models.AttributeID["Status"]), // "1096"
				Values: []models.Value{
					// This assumes you have a selectable status of "Active" in Jira.
					{Value: "Active"},
				},
			},
			{
				ObjectTypeAttributeID: strconv.Itoa(models.AttributeID["Job Role"]), // "1091"
				Values: []models.Value{
					{Value: roleKey},
				},
			},
		},
	}
}

// findEmailInAttributes is a helper function to locate an email value within the Attributes slice.
func findEmailInAttributes(attributes []models.AssetAttribute) string {
	// Get the static ID for the "Email" attribute from our map
	emailAttributeID := strconv.Itoa(models.AttributeID["Email"])

	for _, attr := range attributes {
		if attr.ObjectTypeAttributeID == emailAttributeID {
			if len(attr.Values) > 0 {
				return attr.Values[0].Value
			}
		}
	}
	return ""
}

// saveDataToFile is a helper function to write data to a file for debugging.
func saveDataToFile(filePath string, data interface{}) {
	log.Printf("INFO: Attempting to save data to file: %s", filePath)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("WARN: Failed to marshal data to JSON for saving: %v", err)
		return
	}
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		log.Printf("WARN: Failed to write data to file '%s': %v", filePath, err)
	} else {
		log.Printf("INFO: Data successfully saved to %s", filePath)
	}
}

func safeSubstring(s string, length int) string {
	if len(s) < length {
		return s
	}
	return s[:length]
}
