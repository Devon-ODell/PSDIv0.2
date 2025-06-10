package models

// PaycorConfig holds Paycor API configuration

// This file now accurately models the nested JSON structure from the Paycor API.

// --- Helper Structs for Nested JSON Objects ---

type PositionData struct {
	JobTitle string `json:"jobTitle"`
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
