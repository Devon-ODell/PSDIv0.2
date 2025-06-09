package models

import "encoding/json"

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

// JiraConfig holds Jira API configuration
