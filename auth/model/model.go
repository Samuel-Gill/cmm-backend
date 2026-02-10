package model

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	AccountState string
	CreatedAt    time.Time
}

type RefreshToken struct {
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type PasswordReset struct {
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
}
