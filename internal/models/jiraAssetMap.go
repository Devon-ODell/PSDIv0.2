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
