// internal/jira/jiraClient.go

package jira

import (
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

	apiURL, err := url.Parse(c.cfg.JiraAssetsURL)
	if err != nil {
		// Now using the full URL from config, so this validates it.
		return nil, fmt.Errorf("invalid Jira Assets URL from config: %w", err)
	}

	// Append the specific path for AQL queries
	apiURL = apiURL.JoinPath("aql/objects") // Correct path relative to the v1 base
	// --- END: MODIFICATION ---

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

	// 1. Define a struct that matches the Jira AQL response structure.
	type JiraAQLResponse struct {
		Entries []models.EmployeeAssets `json:"objectEntries"`
	}

	// 2. Unmarshal the response body into this new struct.
	var jiraResponse JiraAQLResponse
	if err := json.Unmarshal(bodyBytes, &jiraResponse); err != nil {
		log.Printf("ERROR: [JiraClient] Failed to unmarshal Jira response: %v", err)
		return nil, fmt.Errorf("failed to unmarshal jira response: %w", err)
	}

	// 3. Return the entries from the parsed response.
	log.Printf("INFO: [JiraClient] Successfully unmarshalled %d employee assets from Jira.", len(jiraResponse.Entries))
	return jiraResponse.Entries, nil
}

// FindObjectsByAQL fetches objects from Jira Assets using a given AQL query.
func (c *Client) FindObjectsByAQL(ctx context.Context, aql string) ([]models.JiraObject, error) {
	queryParams := url.Values{}
	queryParams.Set("aql", aql)
	queryParams.Set("resultsPerPage", "100")

	body, statusCode, err := c.makeAPIRequest(ctx, http.MethodGet, "aql/objects", queryParams, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("Jira API returned non-200 status for AQL query: %d, body: %s", statusCode, string(body))
	}

	var response struct {
		Entries []models.JiraObject `json:"objectEntries"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AQL response: %w. Body: %s", err, string(body))
	}

	return response.Entries, nil
}
