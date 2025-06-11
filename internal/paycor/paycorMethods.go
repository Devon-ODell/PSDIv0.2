package paycor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/Devon-ODell/PSDIv0.2/internal/models"
)

// FetchAllEmployees fetches all employees for the configured LegalEntityID.
func (c *Client) FetchAllEmployees(ctx context.Context) ([]models.Employee, error) {
	if c.cfg.PaycorLegalEntityID == "" {
		return nil, fmt.Errorf("LegalEntityID is not configured in Paycor client")
	}

	var allEmployees []models.Employee
	currentContinuationToken := ""
	apiPath := fmt.Sprintf("/legalentities/%s/employees", c.cfg.PaycorLegalEntityID)
	pageCount := 0

	log.Printf("INFO: [PaycorClient] Starting to fetch all employees for Legal Entity ID: %s", c.cfg.PaycorLegalEntityID)

	for {
		pageCount++
		queryParams := url.Values{}
		if currentContinuationToken != "" {
			queryParams.Set("continuationToken", currentContinuationToken)
		}
		queryParams.Set("include", "All")

		log.Printf("DEBUG: [PaycorClient] Fetching page %d for employees (LE ID %s) with token: %s...",
			pageCount, c.cfg.PaycorLegalEntityID, safeSubstring(currentContinuationToken, 10))

		empBody, _, err := c.makeAPIRequest(ctx, "GET", apiPath, queryParams, nil)
		if err != nil {
			return nil, fmt.Errorf("API call for employees page %d (LE ID %s) failed: %w", pageCount, c.cfg.PaycorLegalEntityID, err)
		}

		var empResponse EmployeesAPIResponse
		if err := json.Unmarshal(empBody, &empResponse); err != nil {
			log.Printf("ERROR: [PaycorClient] Could not unmarshal Employees page %d response for LE ID %s. Raw response snippet:\n%s. Error: %v",
				pageCount, c.cfg.PaycorLegalEntityID, safeSubstring(string(empBody), 500), err)
			return nil, fmt.Errorf("unmarshaling employees response for page %d (LE ID %s): %w", pageCount, c.cfg.PaycorLegalEntityID, err)
		}

		if len(empResponse.Records) > 0 {
			allEmployees = append(allEmployees, empResponse.Records...)
			log.Printf("INFO: [PaycorClient] Fetched %d employees this page (%d total) for LE ID %s.",
				len(empResponse.Records), len(allEmployees), c.cfg.PaycorLegalEntityID)
		} else {
			log.Printf("INFO: [PaycorClient] Fetched 0 employees on page %d for LE ID %s. This might indicate end of data or an issue.", pageCount, c.cfg.PaycorLegalEntityID)
		}

		if empResponse.ContinuationToken != "" {
			currentContinuationToken = empResponse.ContinuationToken
		} else {
			log.Printf("INFO: [PaycorClient] No more continuationToken for LE ID %s after page %d. Finished fetching.", c.cfg.PaycorLegalEntityID, pageCount)
			break
		}
	}

	log.Printf("INFO: [PaycorClient] Successfully fetched a total of %d employees for Legal Entity ID %s over %d pages.", len(allEmployees), c.cfg.PaycorLegalEntityID, pageCount)
	return allEmployees, nil
}

func safeSubstring(s string, length int) string {
	if len(s) < length {
		return s
	}
	return s[:length]
}
