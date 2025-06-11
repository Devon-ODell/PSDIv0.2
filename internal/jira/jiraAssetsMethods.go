package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/Devon-ODell/PSDIv0.2/internal/models"
)

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

// FindOrCreateRole finds a Role asset by name. If it doesn't exist, it creates it.
// Returns the ObjectKey of the found or created role.
func (c *Client) FindOrCreateRole(ctx context.Context, roleName string) (string, error) {
	if roleName == "" {
		return "", nil // Nothing to do if the role name is empty.
	}

	aql := fmt.Sprintf(`objectType = "%s" AND Name = "%s"`, c.cfg.JiraRoleObjectTypeName, roleName)
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
func (c *Client) CreateRoleAsset(ctx context.Context, roleName string) (*models.JiraObject, error) {
	// The "Name" attribute ID for a Role object might be different from an Employee's.
	// You must find this in your Jira Schema configuration. Using a placeholder.
	const roleNameAttributeID = "78" // VERIFY THIS ID

	attributes := []models.AssetAttribute{
		{ObjectTypeAttributeID: roleNameAttributeID, Values: []models.Value{{Value: roleName}}},
	}
	return c.createObject(ctx, c.cfg.JiraRoleObjectTypeID, attributes)
}

// CreateEmployeeAsset creates a new Employee asset.
func (c *Client) CreateEmployeeAsset(ctx context.Context, assetData models.EmployeeAssets) (*models.JiraObject, error) {
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

	_, statusCode, err := c.makeAPIRequest(ctx, http.MethodPut, path, nil, bodyBytes)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("Jira API returned non-200 status on update: %d", statusCode)
	}
	return nil
}

// createObject is a generic helper to create any type of asset object.
func (c *Client) createObject(ctx context.Context, objectTypeID string, attributes []models.AssetAttribute) (*models.JiraObject, error) {
	reqBody := map[string]interface{}{
		"objectTypeId": objectTypeID,
		"attributes":   attributes,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create request body: %w", err)
	}

	respBody, statusCode, err := c.makeAPIRequest(ctx, http.MethodPost, "object/create", nil, bodyBytes)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusCreated {
		return nil, fmt.Errorf("Jira API returned non-201 status on create: %d, body: %s", statusCode, string(respBody))
	}

	var newObject models.JiraObject
	if err := json.Unmarshal(respBody, &newObject); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create response: %w. Body: %s", err, string(respBody))
	}
	log.Printf("SUCCESS: [JiraMethods] Successfully created object with key %s.", newObject.ObjectKey)
	return &newObject, nil
}
