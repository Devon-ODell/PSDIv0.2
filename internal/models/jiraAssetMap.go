package models

// EmployeeAssets represents a single employee record in Jira Assets.
type EmployeeAssets struct {
	ID         string           `json:"id,omitempty"`
	Label      string           `json:"label,omitempty"`
	ObjectType string           `json:"objectType,omitempty"`
	Attributes []AssetAttribute `json:"attributes"`
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
	"Key":                  1085,
	"Name":                 1086,
	"Created":              1087,
	"Updated":              1088,
	"Atlassian Account ID": 1089,
	"Manager Name":         1090,
	"Job Role":             1091, //Object reference field, not simple text
	"Dept":                 1092,
	"Email":                1093,
	"Employment Type":      1094,
	"Start Date":           1095,
	"Status":               1096,
	"Employee Status":      1097,
}
