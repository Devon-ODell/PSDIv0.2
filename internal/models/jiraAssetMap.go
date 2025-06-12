package models

// EmployeeAssets represents a single employee record in Jira Assets.
type EmployeeAssets struct {
	ID         string           `json:"id,omitempty"`
	Label      string           `json:"label,omitempty"`
	ObjectKey  string           `json:"objectKey,omitempty"`
	ObjectType ObjectTypeInfo   `json:"objectType,omitempty"`
	Attributes []AssetAttribute `json:"attributes"`
}

// 1. ADD this new struct definition. You can place it right above EmployeeAssets.
type ObjectTypeInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AssetAttribute represents a key-value pair for an asset's attribute.
type AssetAttribute struct {
	ObjectTypeAttributeID string  `json:"objectTypeAttributeId"`
	Values                []Value `json:"objectAttributeValues"`
}

// Value holds the actual data for an attribute.
type Value struct {
	Value string `json:"value"`
}

// NOTE: These IDs are specific to YOUR Jira instance and schema.
var AttributeID = map[string]int{
	"Key":                  81,
	"Name":                 82,
	"Created":              83,
	"Updated":              84,
	"Atlassian Account ID": 85,
	"Manager Name":         86,
	"Job Role":             87,
	"Dept":                 88,
	"Email":                89,
	"Employment Type":      90,
	"Start Date":           91,
	"Status":               92,
	"Employee Status":      93,
}
