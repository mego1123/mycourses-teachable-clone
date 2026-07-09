package models

import (
	"github.com/google/uuid"
	"time"

)

type WebAuthnCredential struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"userId"`
	CredentialID    string             `json:"-"`
	PublicKey       []byte             `json:"-"`
	AttestationType string             `json:"-"`
	Transport       []string           `json:"-"`
	SignCount       uint32             `json:"-"`
	Name            string             `json:"name"`
	CreatedAt       time.Time          `json:"createdAt"`
	LastUsedAt      *time.Time         `json:"lastUsedAt,omitempty"`
}
