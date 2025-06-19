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
