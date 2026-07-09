package models

import (
	"github.com/google/uuid"
	"time"

)

type TokenType string

const (
	TokenTypeEmailVerification TokenType = "email_verification"
	TokenTypePasswordReset     TokenType = "password_reset"
	TokenTypeMagicLink         TokenType = "magic_link"
	TokenTypeMFA               TokenType = "mfa_pending"
)

type VerificationToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Token     string             `json:"-"`
	Type      TokenType          `json:"type"`
	ExpiresAt time.Time          `json:"expiresAt"`
	CreatedAt time.Time          `json:"createdAt"`
	UsedAt    *time.Time         `json:"usedAt,omitempty"`
}

type RefreshToken struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"userId"`
	TokenHash    string             `json:"-"`
	FamilyID     string             `json:"familyId,omitempty"`
	IPAddress    string             `json:"ipAddress,omitempty"`
	UserAgent    string             `json:"userAgent,omitempty"`
	DeviceInfo   string             `json:"deviceInfo,omitempty"`
	ExpiresAt    time.Time          `json:"expiresAt"`
	CreatedAt    time.Time          `json:"createdAt"`
	LastActiveAt time.Time          `json:"lastActiveAt"`
	IsRevoked    bool               `json:"isRevoked"`
}

type RevokedToken struct {
	ID        uuid.UUID `json:"id"`
	TokenHash string             
	ExpiresAt time.Time          
	CreatedAt time.Time          
}

type OAuthState struct {
	ID        uuid.UUID 
	State     string             
	ExpiresAt time.Time          
	CreatedAt time.Time          
}

type AuthCodeTokenData struct {
	AccessToken  string 
	RefreshToken string 
	MFAToken     string 
	IsMFA        bool   
}

type AuthCode struct {
	ID        uuid.UUID 
	Code      string             
	UserID    uuid.UUID 
	TokenData AuthCodeTokenData  
	ExpiresAt time.Time          
	UsedAt    *time.Time         
	CreatedAt time.Time          
}
