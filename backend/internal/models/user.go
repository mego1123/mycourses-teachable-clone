package models

import (
	"github.com/google/uuid"
	"time"

)

type AuthMethod string

const (
	AuthMethodPassword  AuthMethod = "password"
	AuthMethodGoogle    AuthMethod = "google"
	AuthMethodGitHub    AuthMethod = "github"
	AuthMethodMicrosoft AuthMethod = "microsoft"
	AuthMethodMagicLink AuthMethod = "magic_link"
	AuthMethodPasskey   AuthMethod = "passkey"
)

type User struct {
	ID                   uuid.UUID `json:"id"`
	Email                string             `json:"email" validate:"required,email,max=254"`
	DisplayName          string             `json:"displayName" validate:"required,min=1,max=200"`
	PasswordHash         string             `json:"-"`
	GoogleID             string             `json:"-"`
	GitHubID             string             `json:"-"`
	MicrosoftID          string             `json:"-"`
	AuthMethods          []AuthMethod       `json:"authMethods" validate:"required,min=1,dive,valid_auth_method"`
	EmailVerified        bool               `json:"emailVerified"`
	IsActive             bool               `json:"isActive"`
	TOTPSecret           string             `json:"-"`
	TOTPEnabled          bool               `json:"totpEnabled"`
	TOTPVerifiedAt       *time.Time         `json:"-"`
	RecoveryCodes        []string           `json:"-"`
	ThemePreference      string             `json:"themePreference" validate:"omitempty,oneof=light dark system"`
	OnboardingCompletedAt *time.Time        `json:"onboardingCompletedAt,omitempty"`
	CreatedAt            time.Time          `json:"createdAt" validate:"required"`
	UpdatedAt            time.Time          `json:"updatedAt" validate:"required"`
	LastLoginAt          *time.Time         `json:"lastLoginAt,omitempty"`
	LastVerificationSent *time.Time         `json:"-"`
	FailedLoginAttempts  int                `json:"-"`
	AccountLockedUntil   *time.Time         `json:"-"`
	TrialUsedAt          *time.Time         `json:"trialUsedAt,omitempty"`
}

func (u *User) HasAuthMethod(method AuthMethod) bool {
	for _, m := range u.AuthMethods {
		if m == method {
			return true
		}
	}
	return false
}

func (u *User) IsLocked() bool {
	if u.AccountLockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.AccountLockedUntil)
}
