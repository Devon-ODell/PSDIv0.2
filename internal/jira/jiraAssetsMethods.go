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

	"github.com/Devon-ODell/PSDIv0.2/internal/models"
)

// makeAPIRequest is a generic helper to make authenticated requests to the Jira Assets API.
func (c *Client) makeAPIRequest(ctx context.Context, method, path string, queryParams url.Values, body io.Reader) ([]byte, int, error) {
	apiURL, err := url.Parse(c.cfg.JiraAssetsURL)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid Jira Assets URL from config: %w", err)
	}

	apiURL = apiURL.JoinPath(path)
	if queryParams != nil {
		apiURL.RawQuery = queryParams.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, apiURL.String(), body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create Jira API request: %w", err)
	}

	req.SetBasicAuth(c.cfg.JiraAdminEmail, c.cfg.JiraOrgAPIKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	log.Printf("INFO: [JiraClient] Making %s request to: %s", method, apiURL.String())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute Jira API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("ERROR: [JiraClient] Jira API returned non-2xx status: %s, body: %s", resp.Status, string(bodyBytes))
		return nil, resp.StatusCode, fmt.Errorf("Jira API returned non-2xx status: %s", resp.Status)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read Jira API response body: %w", err)
	}

	return responseBody, resp.StatusCode, nil
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
func (c *Client) FindObjectsByAQL(ctx context.Context, aql string) ([]models.EmployeeAssets, error) {
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
		Entries []models.EmployeeAssets `json:"objectEntries"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AQL response: %w. Body: %s", err, string(body))
	}

	return response.Entries, nil
}

func (c *Client) FindOrCreateRole(ctx context.Context, roleName string) (string, error) {
	if roleName == "" {
		return "", nil
	}

	// Ensure this line uses "Label" as shown below
	aql := fmt.Sprintf(`objectType = "%s" AND Label = "%s"`, c.cfg.JiraRoleObjectTypeName, roleName)

	existingRoles, err := c.FindObjectsByAQL(ctx, aql)
	if err != nil {
		return "", fmt.Errorf("error searching for role '%s': %w", roleName, err)
	}

	if len(existingRoles) > 0 {
		log.Printf("INFO: [JiraMethods] Found existing role '%s' with key %s", roleName, existingRoles[0].ObjectKey)
		return existingRoles[0].ObjectKey, nil
	}

	log.Printf("INFO: [JiraMethods] Role '%s' not found, creating new asset.", roleName)
	newRole, err := c.CreateRoleAsset(ctx, roleName)
	if err != nil {
		return "", fmt.Errorf("failed to create new role asset for '%s': %w", roleName, err)
	}
	return newRole.ObjectKey, nil
}

// CreateRoleAsset creates a new Role asset.
func (c *Client) CreateRoleAsset(ctx context.Context, roleName string) (*models.EmployeeAssets, error) {
	// The "Name" attribute ID for a Role object might be different from an Employee's.
	// You must find this in your Jira Schema configuration. Using a placeholder.
	const roleNameAttributeID = "78" // VERIFY THIS ID

	attributes := []models.AssetAttribute{
		{ObjectTypeAttributeID: roleNameAttributeID, Values: []models.Value{{Value: roleName}}},
	}
	return c.createObject(ctx, c.cfg.JiraRoleObjectTypeID, attributes)
}

// CreateEmployeeAsset creates a new Employee asset.
func (c *Client) CreateEmployeeAsset(ctx context.Context, assetData models.EmployeeAssets) (*models.EmployeeAssets, error) {
	return c.createObject(ctx, c.cfg.JiraEmployeeObjectTypeID, assetData.Attributes)
}

// UpdateEmployeeAsset updates an existing Employee asset in Jira.
func (c *Client) UpdateEmployeeAsset(ctx context.Context, objectID string, assetData models.EmployeeAssets) error {
	path := fmt.Sprintf("object/%s", objectID)
	reqBody := map[string]interface{}{"attributes": assetData.Attributes}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal update request body: %w", err)
	}

	_, statusCode, err := c.makeAPIRequest(ctx, http.MethodPut, path, nil, bytes.NewReader(bodyBytes))

	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("Jira API returned non-200 status on update: %d", statusCode)
	}
	return nil
}

// createObject is a generic helper to create any type of asset object.
func (c *Client) createObject(ctx context.Context, objectTypeID string, attributes []models.AssetAttribute) (*models.EmployeeAssets, error) {
	reqBody := map[string]interface{}{
		"objectTypeId": objectTypeID,
		"attributes":   attributes,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create request body: %w", err)
	}

	respBody, statusCode, err := c.makeAPIRequest(ctx, http.MethodPost, "object/create", nil, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusCreated {
		return nil, fmt.Errorf("Jira API returned non-201 status on create: %d, body: %s", statusCode, string(respBody))
	}

	var newObject models.EmployeeAssets
	if err := json.Unmarshal(respBody, &newObject); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create response: %w. Body: %s", err, string(respBody))
	}
	log.Printf("SUCCESS: [JiraMethods] Successfully created object with key %s.", newObject.ObjectKey)
	return &newObject, nil
}
