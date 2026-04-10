// Package models defines the domain entities for the fantasy football league.
package models

import "time"

// User represents a registered user in the system.
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash *string   `json:"-"` // Never expose in JSON — nullable for OAuth users
	DisplayName  string    `json:"display_name"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	AuthProvider string    `json:"auth_provider"` // "email", "google", "apple"
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateUserRequest is the payload for user registration.
type CreateUserRequest struct {
	Username    string `json:"username"    binding:"required,min=3,max=50"`
	Email       string `json:"email"       binding:"required,email"`
	Password    string `json:"password"    binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
}
