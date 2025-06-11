// internal/jira/jiraClient.go

package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Devon-ODell/PSDIv0.2/internal/config"
	"github.com/Devon-ODell/PSDIv0.2/internal/models"
)

// Client manages communication with the Jira API.
type Client struct {
	cfg        config.JiraConfig
	httpClient *http.Client
}

// NewClient creates a new Jira API client.
func NewClient(cfg config.JiraConfig) (*Client, error) {
	if cfg.JiraAdminEmail == "" || cfg.JiraOrgAPIKey == "" || cfg.JiraSiteName == "" || cfg.JiraWorkspaceID == "" {
		return nil, fmt.Errorf("Jira client configuration is incomplete (Email, API Key, Site Name, Workspace ID are required)")
	}

	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// GetAllEmployeeAssets fetches all objects of the configured Employee type from Jira Assets.
func (c *Client) GetAllEmployeeAssets(ctx context.Context) ([]models.EmployeeAssets, error) {
	// Construct the AQL (Assets Query Language) query to find all "Employee" objects.
	// We use the configured object type name to make it flexible.
	aql := fmt.Sprintf(`objectType = "%s"`, c.cfg.JiraEmployeeObjectTypeName)

	// Jira Assets search endpoint
	apiURL, err := url.Parse(fmt.Sprintf("https://%s", c.cfg.JiraSiteName))
	if err != nil {
		return nil, fmt.Errorf("invalid Jira site name: %w", err)
	}
	apiURL = apiURL.JoinPath(
		"rest/assets/1.0/aql/objects",
	)

	q := apiURL.Query()
	q.Set("aql", aql)
	q.Set("resultsPerPage", "100") // Set a reasonable page size
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jira API request: %w", err)
	}

	// Set required headers for Jira Cloud API
	req.SetBasicAuth(c.cfg.JiraAdminEmail, c.cfg.JiraOrgAPIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	log.Printf("INFO: [JiraClient] Fetching employee assets with AQL: %s", aql)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Jira API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API returned non-200 status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	// Here you would unmarshal the response and map it to your internal models.
	// The Jira Assets API response can be complex. You will need to define structs
	// to match the response and then map the attributes to your `models.EmployeeAsset`.
	// As a placeholder, we'll just log the body.

	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("INFO: [JiraClient] Successfully received data from Jira. Body length: %d bytes", len(bodyBytes))
	log.Printf("DEBUG: [JiraClient] Response Body: %s", string(bodyBytes))

	// TODO: Implement the unmarshalling and mapping from the Jira API response
	// to the []models.EmployeeAsset slice. This will involve creating structs that
	// mirror Jira's JSON response and iterating through the results.

	return []models.EmployeeAssets{}, nil // Return an empty slice for now
}

// Creating an employee asset in Jira
func (c *Client) CreateEmployeeAsset(ctx context.Context, objectTypeId string, employee models.EmployeeAssets) (string, error) { // Returns {
	// Construct the API URL for creating a new object in Jira Assets
	apiURL, err := url.Parse(fmt.Sprintf("https://%s", c.cfg.JiraSiteName))
	if err != nil {
		return "", fmt.Errorf("invalid Jira site name: %w", err)
	}
	apiURL = apiURL.JoinPath(
		"rest/assets/1.0/object",
	)

	reqBody := map[string]interface{}{
		"objectTypeId": objectTypeId,
		"attributes":   employee.Attributes, // Assuming employee has an Attributes field
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL.String(), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create Jira API request: %w", err)
	}

	req.SetBasicAuth(c.cfg.JiraAdminEmail, c.cfg.JiraOrgAPIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	log.Printf("INFO: [JiraClient] Creating employee asset for object type ID: %s", objectTypeId)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute Jira API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Jira API returned non-201 status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var responseData struct {
		ObjectID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return "", fmt.Errorf("failed to decode Jira API response: %w", err)
	}

	log.Printf("INFO: [JiraClient] Successfully created employee asset with ID: %s", responseData.ObjectID)
	return responseData.ObjectID, nil
}

func (c *Client) UpdateEmployeeAsset(ctx context.Context, objectID string, employee models.EmployeeAssets) error {
	// Construct the API URL for updating an existing object in Jira Assets
	apiURL, err := url.Parse(fmt.Sprintf("https://%s", c.cfg.JiraSiteName))
	if err != nil {
		return fmt.Errorf("invalid Jira site name: %w", err)
	}
	apiURL = apiURL.JoinPath(
		"rest/assets/1.0/object", objectID,
	)

	reqBody := map[string]interface{}{
		"attributes": employee.Attributes, // Assuming employee has an Attributes field
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiURL.String(), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create Jira API request: %w", err)
	}

	req.SetBasicAuth(c.cfg.JiraAdminEmail, c.cfg.JiraOrgAPIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	log.Printf("INFO: [JiraClient] Updating employee asset with ID: %s", objectID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute Jira API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Jira API returned non-200 status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	log.Printf("INFO: [JiraClient] Successfully updated employee asset with ID: %s", objectID)
	return nil
}

// Assuming PaycorEmployee is the struct type for data fetched from Paycor.
//func (c *Client) SyncEmployeesToJira(models.Employee, schemaID string, employeeObjectTypeId string) (models.EmployeeAssets, error)
