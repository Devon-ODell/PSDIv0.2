package model

import (
	"time"
)

// PaycorWebhookPayload represents the structure of the webhook payload received from Paycor
type PaycorWebhookPayload struct {
	EventType   string         `json:"eventType"`
	EventID     string         `json:"eventId"`
	EventTime   time.Time      `json:"eventTime"`
	APIClientID string         `json:"apiClientId"`
	Body        PaycorEmployee `json:"body"`
}

// PaycorEmployee represents an employee record from Paycor
type PaycorEmployee struct {
	EmployeeID           string                `json:"employeeId"`
	FirstName            string                `json:"firstName"`
	LastName             string                `json:"lastName"`
	MiddleName           string                `json:"middleName,omitempty"`
	PreferredName        string                `json:"preferredName,omitempty"`
	PersonalEmail        string                `json:"personalEmail,omitempty"`
	WorkEmail            string                `json:"workEmail,omitempty"`
	Department           string                `json:"department,omitempty"`
	Location             string                `json:"location,omitempty"`
	EmploymentDateData   *EmploymentDateData   `json:"employmentDateData,omitempty"`
	PositionData         *PositionData         `json:"positionData,omitempty"`
	StatusData           *StatusData           `json:"statusData,omitempty"`
	CompensationData     *CompensationData     `json:"compensationData,omitempty"`
	EmergencyContactData *EmergencyContactData `json:"emergencyContactData,omitempty"`
}

// EmploymentDateData represents employment date information
type EmploymentDateData struct {
	HireDate         string `json:"hireDate,omitempty"`
	AdjustedHireDate string `json:"adjustedHireDate,omitempty"`
	TerminationDate  string `json:"terminationDate,omitempty"`
	RehireDate       string `json:"rehireDate,omitempty"`
	LastDayWorked    string `json:"lastDayWorked,omitempty"`
	RetirementDate   string `json:"retirementDate,omitempty"`
}

// PositionData represents position information
type PositionData struct {
	JobTitle      string  `json:"jobTitle,omitempty"`
	Manager       string  `json:"manager,omitempty"`
	ManagerID     string  `json:"managerId,omitempty"`
	EmployeeType  string  `json:"employeeType,omitempty"`
	EmployeeClass string  `json:"employeeClass,omitempty"`
	PayGroup      string  `json:"payGroup,omitempty"`
	PayFrequency  string  `json:"payFrequency,omitempty"`
	FLSAStatus    string  `json:"flsaStatus,omitempty"`
	StandardHours float64 `json:"standardHours,omitempty"`
}

// StatusData represents employee status information
type StatusData struct {
	Status              string `json:"status,omitempty"`
	StatusEffectiveDate string `json:"statusEffectiveDate,omitempty"`
	StatusReason        string `json:"statusReason,omitempty"`
}

// CompensationData represents compensation information
type CompensationData struct {
	Rate          float64 `json:"rate,omitempty"`
	RateUnit      string  `json:"rateUnit,omitempty"`
	EffectiveDate string  `json:"effectiveDate,omitempty"`
	EndDate       string  `json:"endDate,omitempty"`
	AnnualAmount  float64 `json:"annualAmount,omitempty"`
	Currency      string  `json:"currency,omitempty"`
}

// EmergencyContactData represents emergency contact information
type EmergencyContactData struct {
	PrimaryContact   *ContactInfo `json:"primaryContact,omitempty"`
	SecondaryContact *ContactInfo `json:"secondaryContact,omitempty"`
}

// ContactInfo represents contact information
type ContactInfo struct {
	Name         string   `json:"name,omitempty"`
	Relationship string   `json:"relationship,omitempty"`
	PhoneNumber  string   `json:"phoneNumber,omitempty"`
	EmailAddress string   `json:"emailAddress,omitempty"`
	Address      *Address `json:"address,omitempty"`
}

// Address represents an address
type Address struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	Country    string `json:"country,omitempty"`
}
