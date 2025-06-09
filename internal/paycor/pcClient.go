package paycor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	// Assuming your config package is correctly pathed if this client is used elsewhere
	// For direct use by cmd/server/main.go, config.PaycorConfig will be passed in.

	"golang.org/x/oauth2"
)

// Custom UnmarshalJSON to handle additional properties
func (e *Employee) UnmarshalJSON(data []byte) error {
	// Temporary struct to unmarshal known fields
	type Alias Employee
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal all data into a map to find extra fields
	var allData map[string]interface{}
	if err := json.Unmarshal(data, &allData); err != nil {
		return err
	}

	// Remove known fields from allData to isolate additional properties
	delete(allData, "id")
	delete(allData, "legalEntityId")
	delete(allData, "firstName")
	delete(allData, "lastName")
	delete(allData, "middleName")
	delete(allData, "preferredName")
	delete(allData, "personalEmail")
	delete(allData, "workEmail")
	delete(allData, "lastModifiedDate")
	delete(allData, "department")
	delete(allData, "locationName")
	delete(allData, "addresses")
	delete(allData, "birthDate")
	delete(allData, "businessUnit")
	delete(allData, "callInData")
	delete(allData, "citizenship")
	delete(allData, "compensation")
	delete(allData, "departmentAndPosition")
	delete(allData, "directDeposit")
	delete(allData, "emergencyContacts")
	delete(allData, "employmentDates")
	delete(allData, "employmentStatus")
	delete(allData, "ethnicityRace")
	delete(allData, "gender")
	delete(allData, "licenses")
	delete(allData, "personIdentification")
	delete(allData, "phones")
	delete(allData, "position")
	delete(allData, "status")
	delete(allData, "taxes")
	delete(allData, "union")
	delete(allData, "veteran")
	delete(allData, "workLocation")

	if len(allData) > 0 {
		e.AdditionalProperties = allData
	}

	return nil
}

// EmployeesAPIResponse models the expected API response structure when fetching employees.
type EmployeesAPIResponse struct {
	Items             []Employee `json:"items"`
	ContinuationToken string     `json:"continuationToken"`
}

// Client manages communication with the Paycor API.
type Client struct {
	paycorCfg   PaycorConfig // Using a direct struct, not from a shared config package for this example
	httpClient  *http.Client
	tokenSource oauth2.TokenSource // Kept for potential direct use, though httpClient manages tokens
}

// PaycorConfig is a local copy of the necessary fields from the main config.
// This makes the paycor client package more self-contained if used independently.
type PaycorConfig struct {
	ClientID        string
	ClientSecret    string
	SubscriptionKey string
	RefreshToken    string
	TokenURLBase    string
	APIBaseURL      string
	LegalEntityID   string
	Scopes          []string
	OutputFilePath  string // Added here if client needs to know about it, though typically main handles output
}

// loggingTokenSource (same as before)
type loggingTokenSource struct {
	src              oauth2.TokenSource
	lastRefreshToken string
	paycorCfg        PaycorConfig // To allow updating the refresh token in config if it changes
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
		// Here you would ideally have a mechanism to update the persisted RefreshToken
		// For this script, we'll just log it. The user needs to manually update their .env
		log.Printf("ACTION REQUIRED: New Refresh Token: %s", token.RefreshToken)
	}
	return token, nil
}

// NewClient creates a new Paycor API client.
// It takes a PaycorConfig struct directly.
func NewClient(ctx context.Context, cfg PaycorConfig) (*Client, error) {
	if cfg.TokenURLBase == "" || cfg.APIBaseURL == "" || cfg.ClientID == "" ||
		cfg.ClientSecret == "" || cfg.SubscriptionKey == "" || cfg.RefreshToken == "" {
		return nil, fmt.Errorf("Paycor client configuration is incomplete in NewClient")
	}

	parsedTokenURL, err := url.Parse(cfg.TokenURLBase)
	if err != nil {
		return nil, fmt.Errorf("invalid Paycor Token URL Base '%s': %w", cfg.TokenURLBase, err)
	}
	queryToken := parsedTokenURL.Query()
	queryToken.Set("subscription-key", cfg.SubscriptionKey)
	parsedTokenURL.RawQuery = queryToken.Encode()
	fullTokenURL := parsedTokenURL.String()

	oauthConf := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       cfg.Scopes,
		Endpoint:     oauth2.Endpoint{TokenURL: fullTokenURL},
	}

	initialToken := &oauth2.Token{
		RefreshToken: cfg.RefreshToken,
		Expiry:       time.Now().Add(-1 * time.Hour), // Force initial refresh
	}

	loggingTS := &loggingTokenSource{src: oauthConf.TokenSource(ctx, initialToken), lastRefreshToken: cfg.RefreshToken, paycorCfg: cfg}

	// Customize HTTP client for oauth2 if needed (e.g., for timeouts)
	customHTTPClient := &http.Client{
		Timeout: 90 * time.Second, // Example: longer timeout for potentially large API calls
	}
	authCtx := context.WithValue(ctx, oauth2.HTTPClient, customHTTPClient)
	authedClient := oauth2.NewClient(authCtx, loggingTS)

	return &Client{
		paycorCfg:   cfg,
		httpClient:  authedClient,
		tokenSource: loggingTS,
	}, nil
}

func (c *Client) makeAPIRequest(ctx context.Context, method, path string, queryParams url.Values, body io.Reader) ([]byte, int, error) {
	fullURL, err := url.Parse(c.paycorCfg.APIBaseURL)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid Paycor API Base URL '%s': %w", c.paycorCfg.APIBaseURL, err)
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

	req.Header.Add("Ocp-Apim-Subscription-Key", c.paycorCfg.SubscriptionKey)
	req.Header.Add("Accept", "application/json")
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	log.Printf("INFO: [PaycorClient] Attempting API %s request to: %s", method, urlStr)
	resp, err := c.httpClient.Do(req) // httpClient already includes Authorization
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

// FetchAllEmployees fetches all employees for the configured LegalEntityID.
func (c *Client) FetchAllEmployees(ctx context.Context) ([]Employee, error) {
	if c.paycorCfg.LegalEntityID == "" {
		return nil, fmt.Errorf("LegalEntityID is not configured in Paycor client")
	}

	var allEmployees []Employee
	currentContinuationToken := ""
	apiPath := fmt.Sprintf("/legalentities/%s/employees", c.paycorCfg.LegalEntityID)
	pageCount := 0

	log.Printf("INFO: [PaycorClient] Starting to fetch all employees for Legal Entity ID: %s", c.paycorCfg.LegalEntityID)

	for {
		pageCount++
		queryParams := url.Values{}
		if currentContinuationToken != "" {
			queryParams.Set("continuationToken", currentContinuationToken)
		}
		// To get all possible data, "All" is a good general choice.
		// Refer to Paycor documentation for specific 'include' values if "All" is too much or misses something.
		// Example: "Addresses,BirthDate,BusinessUnit,CallInData,Citizenship,Compensation,DepartmentAndPosition,DirectDeposit,EmergencyContacts,EmploymentDates,EmploymentStatus,EthnicityRace,Gender,Licenses,PersonIdentification,Phones,Position,Status,Taxes,Union,Veteran,WorkLocation"
		queryParams.Set("include", "All")
		queryParams.Set("pageSize", "100") // Request a larger page size if API supports

		log.Printf("DEBUG: [PaycorClient] Fetching page %d for employees (LE ID %s) with token: %s...",
			pageCount, c.paycorCfg.LegalEntityID, safeSubstring(currentContinuationToken, 10))

		empBody, _, err := c.makeAPIRequest(ctx, "GET", apiPath, queryParams, nil)
		if err != nil {
			return nil, fmt.Errorf("API call for employees page %d (LE ID %s) failed: %w", pageCount, c.paycorCfg.LegalEntityID, err)
		}

		var empResponse EmployeesAPIResponse
		if err := json.Unmarshal(empBody, &empResponse); err != nil {
			log.Printf("ERROR: [PaycorClient] Could not unmarshal Employees page %d response for LE ID %s. Raw response snippet:\n%s. Error: %v",
				pageCount, c.paycorCfg.LegalEntityID, safeSubstring(string(empBody), 500), err)
			return nil, fmt.Errorf("unmarshaling employees response for page %d (LE ID %s): %w", pageCount, c.paycorCfg.LegalEntityID, err)
		}

		if len(empResponse.Items) > 0 {
			allEmployees = append(allEmployees, empResponse.Items...)
			log.Printf("INFO: [PaycorClient] Fetched %d employees this page (%d total) for LE ID %s.",
				len(empResponse.Items), len(allEmployees), c.paycorCfg.LegalEntityID)
		} else {
			log.Printf("INFO: [PaycorClient] Fetched 0 employees on page %d for LE ID %s. This might indicate end of data or an issue.", pageCount, c.paycorCfg.LegalEntityID)
		}

		if empResponse.ContinuationToken != "" {
			currentContinuationToken = empResponse.ContinuationToken
		} else {
			log.Printf("INFO: [PaycorClient] No more continuationToken for LE ID %s after page %d. Finished fetching.", c.paycorCfg.LegalEntityID, pageCount)
			break
		}
	}

	log.Printf("INFO: [PaycorClient] Successfully fetched a total of %d employees for Legal Entity ID %s over %d pages.", len(allEmployees), c.paycorCfg.LegalEntityID, pageCount)
	return allEmployees, nil
}

func safeSubstring(s string, length int) string {
	if len(s) < length {
		return s
	}
	return s[:length]
}
