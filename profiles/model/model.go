package model

import "time"

type Profile struct {
	UserID          string    `json:"user_id"`
	Age             int       `json:"age"`
	Profession      string    `json:"profession"`
	Education       string    `json:"education"`
	Income          int       `json:"income"`
	ResidencyStatus string    `json:"residency_status"`
	Location        string    `json:"location"`
	MaritalStatus   string    `json:"marital_status"`
	Description     string    `json:"description"`
	HideContactInfo bool      `json:"hide_contact_info"`
	HideAddress     bool      `json:"hide_address"`
	HideIncome      bool      `json:"hide_income"`
	HideVisaStatus  bool      `json:"hide_visa_status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ListResult struct {
	Items []Profile `json:"items"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
}
