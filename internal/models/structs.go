package models

import "encoding/json"

// PaycorConfig holds Paycor API configuration

// This file now accurately models the nested JSON structure from the Paycor API.

// --- Helper Structs for Nested JSON Objects ---

type PositionData struct {
	JobTitle string `json:"jobTitle"`
	Manager  string `json:"manager,omitempty"`
}

type Email struct {
	Type         string `json:"type"`
	EmailAddress string `json:"emailAddress"`
}

type EmploymentDateData struct {
	HireDate        string `json:"hireDate"`
	TerminationDate string `json:"terminationDate"`
}

type StatusData struct {
	Status string `json:"status"`
}

type WorkLocation struct {
	Name  string `json:"name"`
	City  string `json:"city"`
	State string `json:"state"`
}

type LegalEntity struct {
	ID string `json:"id"`
}

// Employee struct is designed to be flexible for "include=All".
// --- Main Employee Struct ---

type Employee struct {
	ID                 string             `json:"id"`
	FirstName          string             `json:"firstName"`
	LastName           string             `json:"lastName"`
	EmployeeNumber     string             `json:"employeeNumber"`
	Email              Email              `json:"email"`
	PositionData       PositionData       `json:"positionData"`
	EmploymentDateData EmploymentDateData `json:"employmentDateData"`
	StatusData         StatusData         `json:"statusData"`
	WorkLocation       WorkLocation       `json:"workLocation"`
	LegalEntity        LegalEntity        `json:"legalEntity"`
}

// JiraConfig holds Jira API configuration

// JiraIssueRequest is the top-level struct for creating a Jira issue.
type JiraIssueRequest struct {
	Fields JiraIssueFields `json:"fields"`
}

// JiraIssueFields contains all the fields for a new issue.
type JiraIssueFields struct {
	Project      JiraProject            `json:"project"`
	Summary      string                 `json:"summary"`
	Description  JiraIssueDescription   `json:"description,omitempty"`
	IssueType    JiraIssueType          `json:"issuetype"`
	CustomFields map[string]interface{} `json:"-"` // This will be handled dynamically
}

// MarshalJSON is a custom marshaller to include the dynamic custom fields.
func (f JiraIssueFields) MarshalJSON() ([]byte, error) {
	// Use an alias to avoid recursion
	type Alias JiraIssueFields

	// Start with the standard fields
	m := map[string]interface{}{
		"project":   f.Project,
		"summary":   f.Summary,
		"issuetype": f.IssueType,
	}

	if f.Description.Content != nil {
		m["description"] = f.Description
	}

	// Add the custom fields to the map
	for key, value := range f.CustomFields {
		m[key] = value
	}

	return json.Marshal(m)
}

// JiraProject identifies the project by its key.
type JiraProject struct {
	Key string `json:"key"`
}

// JiraIssueType identifies the issue type by its name.
type JiraIssueType struct {
	Name string `json:"name"`
}

// JiraIssueDescription represents the rich text description field.
type JiraIssueDescription struct {
	Type    string                   `json:"type"`
	Version int                      `json:"version"`
	Content []JiraDescriptionContent `json:"content"`
}

// JiraDescriptionContent is part of the rich text format.
type JiraDescriptionContent struct {
	Type    string                `json:"type"`
	Content []JiraDescriptionText `json:"content"`
}

// JiraDescriptionText is the final text part of the rich text format.
type JiraDescriptionText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// JiraIssueResponse models the response from Jira after creating an issue.
type JiraIssueResponse struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}
