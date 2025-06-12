// internal/jira/jiraClient.go

package jira

import (
	/* 	"context"
	   	"encoding/json" */
	"fmt"
	/* "io"
	"log"
	"net/url" */
	"net/http"

	"time"

	"github.com/Devon-ODell/PSDIv0.2/internal/config"
	/* "github.com/Devon-ODell/PSDIv0.2/internal/models" */)

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

// makeAPIRequest is a generic helper to make authenticated requests to the Jira Assets API.
/* func (c *Client) makeAPIRequest(ctx context.Context, method, path string, queryParams url.Values, body io.Reader) ([]byte, int, error) {
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
*/
// GetAllEmployeeAssets fetches all objects of the configured Employee type from Jira Assets.
