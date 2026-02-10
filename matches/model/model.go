package model

import "time"

type BrowseFilter struct {
	MinAge          int
	MaxAge          int
	Gender          string
	MinIncome       int
	MaxIncome       int
	Location        string
	ResidencyStatus string
	Page            int
	Size            int
}

type BrowseProfile struct {
	UserID          string    `json:"user_id"`
	Age             int       `json:"age"`
	Gender          string    `json:"gender"`
	Profession      string    `json:"profession"`
	Education       string    `json:"education"`
	Income          int       `json:"income"`
	Location        string    `json:"location"`
	ResidencyStatus string    `json:"residency_status"`
	MaritalStatus   string    `json:"marital_status"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

type BrowseResult struct {
	Items []BrowseProfile `json:"items"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

type RelationshipSummary struct {
	LikedBy       []string `json:"liked_by"`
	Liked         []string `json:"liked"`
	MutualMatches []string `json:"mutual_matches"`
}
