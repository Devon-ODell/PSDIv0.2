// internal/paycor/paycorClient.go

package paycor

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	// Import the central config package
	"github.com/Devon-ODell/PSDIv0.2/internal/config"
	"github.com/Devon-ODell/PSDIv0.2/internal/models"
	"golang.org/x/oauth2"
)

// NOTE: The UnmarshalJSON method for models.Employee has been moved to internal/models/structs.go
// This is required because methods can only be defined on types within the same package.

// EmployeesAPIResponse models the expected API response structure when fetching employees.
type EmployeesAPIResponse struct {
	Records           []models.Employee `json:"records"`
	ContinuationToken string            `json:"continuationToken"`
}

// Client manages communication with the Paycor API.
type Client struct {
	cfg        config.PaycorConfig // Use the imported config struct
	httpClient *http.Client
}

// loggingTokenSource (same as before, but references the central config)
type loggingTokenSource struct {
	src              oauth2.TokenSource
	lastRefreshToken string
	paycorCfg        config.PaycorConfig // Use the imported config struct
}

func (s *loggingTokenSource) Token() (*oauth2.Token, error) {
	log.Println("DEBUG: [PaycorTokenSource] Attempting to retrieve/refresh token...")
	token, err := s.src.Token()
	if err != nil {
		log.Printf("ERROR: [PaycorTokenSource] Failed to retrieve/refresh token: %v", err)
		if retrieveError, ok := err.(*oauth2.RetrieveError); ok {
			log.Printf("DEBUG: [PaycorTokenSource] OAuth2 RetrieveError details:")
			if retrieveError.Response != nil {
				log.Printf("  HTTP Status Code: %d", retrieveError.Response.StatusCode)
			}
			log.Printf("  Response Body: %s", string(retrieveError.Body))
		}
		return nil, err
	}
	log.Printf("DEBUG: [PaycorTokenSource] Successfully retrieved/refreshed token.")
	log.Printf("  Expires At (UTC): %s", token.Expiry.UTC().Format(time.RFC3339))

	if token.RefreshToken != "" && token.RefreshToken != s.lastRefreshToken {
		log.Printf("INFO: [PaycorTokenSource] A new Refresh Token was issued (masked): %s...", safeSubstring(token.RefreshToken, 10))
		log.Println("INFO: [PaycorTokenSource] IMPORTANT: The new refresh token should be saved securely and used for subsequent runs.")
		s.lastRefreshToken = token.RefreshToken
		log.Printf("ACTION REQUIRED: New Refresh Token: %s", token.RefreshToken)
	}
	return token, nil
}

// NewClient creates a new Paycor API client.
// It now accepts the central config.PaycorConfig struct.
func NewClient(ctx context.Context, cfg config.PaycorConfig) (*Client, error) {
	if cfg.PaycorTokenURLBase == "" || cfg.PaycorAPIBaseURL == "" || cfg.PaycorClientID == "" ||
		cfg.PaycorClientSecret == "" || cfg.PaycorOcpApimSubscriptionKey == "" || cfg.PaycorRefreshToken == "" {
		return nil, fmt.Errorf("Paycor client configuration is incomplete in NewClient")
	}

	parsedTokenURL, err := url.Parse(cfg.PaycorTokenURLBase)
	if err != nil {
		return nil, fmt.Errorf("invalid Paycor Token URL Base '%s': %w", cfg.PaycorTokenURLBase, err)
	}
	queryToken := parsedTokenURL.Query()
	queryToken.Set("subscription-key", cfg.PaycorOcpApimSubscriptionKey)
	parsedTokenURL.RawQuery = queryToken.Encode()
	fullTokenURL := parsedTokenURL.String()

	oauthConf := &oauth2.Config{
		ClientID:     cfg.PaycorClientID,
		ClientSecret: cfg.PaycorClientSecret,
		Scopes:       cfg.PaycorScopes,
		Endpoint:     oauth2.Endpoint{TokenURL: fullTokenURL},
	}

	initialToken := &oauth2.Token{
		RefreshToken: cfg.PaycorRefreshToken,
		Expiry:       time.Now().Add(-1 * time.Hour), // Force initial refresh
	}

	loggingTS := &loggingTokenSource{src: oauthConf.TokenSource(ctx, initialToken), lastRefreshToken: cfg.PaycorRefreshToken, paycorCfg: cfg}

	customHTTPClient := &http.Client{
		Timeout: 90 * time.Second,
	}
	authCtx := context.WithValue(ctx, oauth2.HTTPClient, customHTTPClient)
	authedClient := oauth2.NewClient(authCtx, loggingTS)

	return &Client{
		cfg:        cfg,
		httpClient: authedClient,
	}, nil
}

func (c *Client) makeAPIRequest(ctx context.Context, method, path string, queryParams url.Values, body io.Reader) ([]byte, int, error) {
	fullURL, err := url.Parse(c.cfg.PaycorAPIBaseURL)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid Paycor API Base URL '%s': %w", c.cfg.PaycorAPIBaseURL, err)
	}
	fullURL = fullURL.JoinPath(path)
	if queryParams != nil {
		fullURL.RawQuery = queryParams.Encode()
	}
	urlStr := fullURL.String()

	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request for %s: %w", urlStr, err)
	}

	req.Header.Add("Ocp-Apim-Subscription-key", c.cfg.PaycorOcpApimSubscriptionKey)
	req.Header.Add("Accept", "application/json")
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	log.Printf("INFO: [PaycorClient] Attempting API %s request to: %s", method, urlStr)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("making API request to %s: %w", urlStr, err)
	}
	defer resp.Body.Close()

	log.Printf("INFO: [PaycorClient] API Response Status from %s: %s", urlStr, resp.Status)
	responseBodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading API response body from %s: %w", urlStr, readErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("ERROR: [PaycorClient] API request to %s failed with status %d. Body: %s", urlStr, resp.StatusCode, string(responseBodyBytes))
		return responseBodyBytes, resp.StatusCode, fmt.Errorf("API request to %s failed with status %d. Body: %s", urlStr, resp.StatusCode, string(responseBodyBytes))
	}

	return responseBodyBytes, resp.StatusCode, nil
}
