package models

// PaycorConfig holds Paycor API configuration

// Employee struct is designed to be flexible for "include=All".
// It uses map[string]interface{} for sections where fields can vary greatly
// or are deeply nested and you want to capture everything without defining exhaustive structs.
// Alternatively, you can define very detailed structs for each known section from Paycor docs.
type Employee struct {
	ID               string `json:"id"`
	LegalEntityID    string `json:"legalEntityId,omitempty"` // The LE this employee belongs to
	FirstName        string `json:"firstName,omitempty"`
	LastName         string `json:"lastName,omitempty"`
	MiddleName       string `json:"middleName,omitempty"`
	PreferredName    string `json:"preferredName,omitempty"`
	PersonalEmail    string `json:"personalEmail,omitempty"`
	WorkEmail        string `json:"workEmail,omitempty"`
	LastModifiedDate string `json:"lastModifiedDate,omitempty"` // Example: "2024-01-15T18:09:07.63Z"

	// Common top-level fields often requested
	Department   string `json:"department,omitempty"`   // Often part of Position
	LocationName string `json:"locationName,omitempty"` // Often part of WorkLocation or Position

	// Using map for potentially complex/variable nested structures from "include=All"
	// These correspond to the 'include' options in Paycor's API.
	Addresses             []interface{} `json:"addresses,omitempty"`             // from "Addresses"
	BirthDate             *interface{}  `json:"birthDate,omitempty"`             // from "BirthDate"
	BusinessUnit          *interface{}  `json:"businessUnit,omitempty"`          // from "BusinessUnit"
	CallInData            *interface{}  `json:"callInData,omitempty"`            // from "CallInData"
	Citizenship           *interface{}  `json:"citizenship,omitempty"`           // from "Citizenship"
	Compensation          *interface{}  `json:"compensation,omitempty"`          // from "Compensation"
	DepartmentAndPosition *interface{}  `json:"departmentAndPosition,omitempty"` // from "DepartmentAndPosition" (might be redundant with Position)
	DirectDeposit         []interface{} `json:"directDeposit,omitempty"`         // from "DirectDeposit"
	EmergencyContacts     []interface{} `json:"emergencyContacts,omitempty"`     // from "EmergencyContacts"
	EmploymentDates       *interface{}  `json:"employmentDates,omitempty"`       // from "EmploymentDates"
	EmploymentStatus      *interface{}  `json:"employmentStatus,omitempty"`      // from "EmploymentStatus" (might be redundant with Status)
	EthnicityRace         *interface{}  `json:"ethnicityRace,omitempty"`         // from "EthnicityRace"
	Gender                *interface{}  `json:"gender,omitempty"`                // from "Gender"
	Licenses              []interface{} `json:"licenses,omitempty"`              // from "Licenses"
	PersonIdentification  *interface{}  `json:"personIdentification,omitempty"`  // from "PersonIdentification" (SSN, etc.)
	Phones                []interface{} `json:"phones,omitempty"`                // from "Phones"
	Position              *interface{}  `json:"position,omitempty"`              // from "Position"
	Status                *interface{}  `json:"status,omitempty"`                // from "Status"
	Taxes                 []interface{} `json:"taxes,omitempty"`                 // from "Taxes"
	Union                 *interface{}  `json:"union,omitempty"`                 // from "Union"
	Veteran               *interface{}  `json:"veteran,omitempty"`               // from "Veteran"
	WorkLocation          *interface{}  `json:"workLocation,omitempty"`          // from "WorkLocation"

	// Catch-all for any other properties not explicitly defined
	AdditionalProperties map[string]interface{} `json:"-"` // Use json.RawMessage or custom unmarshal if needed
}

// JiraConfig holds Jira API configuration
