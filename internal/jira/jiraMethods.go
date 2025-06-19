// internal/jira/jiraMethods.go

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

// --- NEW METHODS FOR STANDARD JIRA API ---

// makeStandardAPIRequest is a generic helper for the standard v3 Jira Cloud API.
// It uses a different base URL than the Assets API.
func (c *Client) makeStandardAPIRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, int, error) {
	// Construct the URL for the standard Jira Cloud API (e.g., https://your-domain.atlassian.net/rest/api/3)
	fullURL, err := url.Parse(fmt.Sprintf("https://%s", c.cfg.JiraSiteName))
	if err != nil {
		return nil, 0, fmt.Errorf("invalid Jira Site Name from config: %w", err)
	}
	fullURL = fullURL.JoinPath("rest", "api", "3", path)

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create standard Jira API request: %w", err)
	}

	req.SetBasicAuth(c.cfg.JiraAdminEmail, c.cfg.JiraOrgAPIKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	log.Printf("INFO: [JiraClient] Making %s request to standard API: %s", method, fullURL.String())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute standard Jira API request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read standard Jira API response body: %w", readErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("ERROR: [JiraClient] Standard Jira API returned non-2xx status: %s, body: %s", resp.Status, string(responseBody))
		return responseBody, resp.StatusCode, fmt.Errorf("standard Jira API returned non-2xx status: %s", resp.Status)
	}

	return responseBody, resp.StatusCode, nil
}

// CreateIssueWithAsset creates a new Jira issue and links it to an asset.
func (c *Client) CreateIssueWithAsset(ctx context.Context, projectKey, summary, description, assetCustomFieldID, assetObjectKey string) (*models.JiraIssueResponse, error) {

	// Construct the payload for the Jira issue.
	// The structure must match the Jira API format exactly.
	issuePayload := models.JiraIssueRequest{
		Fields: models.JiraIssueFields{
			Project: models.JiraProject{
				Key: projectKey,
			},
			Summary: summary,
			IssueType: models.JiraIssueType{
				Name: "Task", // Or use a configurable value from cfg
			},
			Description: models.JiraIssueDescription{
				Type:    "doc",
				Version: 1,
				Content: []models.JiraDescriptionContent{
					{
						Type: "paragraph",
						Content: []models.JiraDescriptionText{
							{
								Type: "text",
								Text: description,
							},
						},
					},
				},
			},
			// This is how you set the custom field for the Asset object.
			// The key must be the custom field ID, e.g., "customfield_10050".
			CustomFields: map[string]interface{}{
				assetCustomFieldID: []string{assetObjectKey},
			},
		},
	}

	// Marshal the payload into JSON.
	bodyBytes, err := json.Marshal(issuePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal issue creation payload: %w", err)
	}
	log.Printf("DEBUG: [JiraClient] Issue Creation Payload: %s", string(bodyBytes))

	// Make the API call to create the issue.
	respBody, _, err := c.makeStandardAPIRequest(ctx, http.MethodPost, "issue", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	// Unmarshal the response from Jira.
	var issueResponse models.JiraIssueResponse
	if err := json.Unmarshal(respBody, &issueResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal issue creation response: %w. Body: %s", err, string(respBody))
	}

	return &issueResponse, nil
}
