package models

type JiraAssetMap struct {
	PaycorSystemID   string `json:"paycor_system_id"`   // Unique Paycor Employee GUID
	PaycorEmployeeID string `json:"paycor_employee_id"` // Company-assigned employee number
	FirstName        string `json:"first_name"`         // Employee's first name
	LastName         string `json:"last_name"`          // Employee's last name
	Email            string `json:"email"`              // Employee's work email
	Department       string `json:"department"`         // Employee's department
	JobTitle         string `json:"job_title"`          // Employee's job title
	Location         string `json:"location"`           // Employee's work location
	StartDate        string `json:"start_date"`         // Employee's hire date
}
