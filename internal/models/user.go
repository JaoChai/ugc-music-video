package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email            string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash     string    `json:"-" gorm:"not null"`
	Name             *string   `json:"name"`
	OpenRouterModel  string    `json:"openrouter_model" gorm:"default:''"`
	OpenRouterAPIKey *string   `json:"-"` // Encrypted, never expose in JSON
	KIEAPIKey        *string   `json:"-"` // Encrypted, never expose in JSON
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CreateUserInput represents the input for user registration
type CreateUserInput struct {
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=8"`
	Name     *string `json:"name"`
}

// LoginInput represents the input for user login
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateUserInput represents the input for updating user profile
type UpdateUserInput struct {
	Name            *string `json:"name"`
	OpenRouterModel *string `json:"openrouter_model"`
}

// UpdateAPIKeysInput represents the input for updating user API keys
type UpdateAPIKeysInput struct {
	OpenRouterAPIKey *string `json:"openrouter_api_key"`
	KIEAPIKey        *string `json:"kie_api_key"`
}

// APIKeysStatusResponse represents the API keys status (not actual keys)
type APIKeysStatusResponse struct {
	HasOpenRouterKey bool `json:"has_openrouter_key"`
	HasKIEKey        bool `json:"has_kie_key"`
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID              uuid.UUID `json:"id"`
	Email           string    `json:"email"`
	Name            *string   `json:"name"`
	OpenRouterModel string    `json:"openrouter_model"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ToResponse converts a User to UserResponse (excludes sensitive data)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		Name:            u.Name,
		OpenRouterModel: u.OpenRouterModel,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}
