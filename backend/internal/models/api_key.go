package models

import (
	"github.com/google/uuid"
	"time"

)

type APIKeyAuthority string

const (
	APIKeyAuthorityAdmin APIKeyAuthority = "admin"
	APIKeyAuthorityUser  APIKeyAuthority = "user"
)

func ValidAPIKeyAuthority(a APIKeyAuthority) bool {
	return a == APIKeyAuthorityAdmin || a == APIKeyAuthorityUser
}

type APIKey struct {
	ID         uuid.UUID `json:"id"`
	Name       string             `json:"name" validate:"required,min=1,max=100"`
	KeyHash    string             `json:"-" validate:"required"`
	KeyPreview string             `json:"keyPreview" validate:"required"`
	Authority  APIKeyAuthority    `json:"authority" validate:"required,valid_api_authority"`
	CreatedBy  uuid.UUID `json:"createdBy" validate:"required"`
	CreatedAt  time.Time          `json:"createdAt" validate:"required"`
	LastUsedAt *time.Time         `json:"lastUsedAt"`
	IsActive   bool               `json:"isActive"`
}
