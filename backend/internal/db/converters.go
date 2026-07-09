// Package db — Type converters between sqlc-generated types and domain models.
// Maps between gen.* (Postgres, snake_case) and models.* (domain, camelCase).
package db

import (
	"github.com/google/uuid"

	gen "mycourses/internal/db/gen"
	"mycourses/internal/models"
)

// ToUser converts a sqlc-generated gen.User to a models.User.
func ToUser(u gen.User) models.User {
	user := models.User{
		ID:            u.ID,
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		EmailVerified: u.EmailVerified,
		IsActive:      u.IsActive,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
	if u.PasswordHash != nil {
		user.PasswordHash = *u.PasswordHash
	}
	if u.GoogleID != nil {
		user.GoogleID = *u.GoogleID
	}
	if u.GithubID != nil {
		user.GitHubID = *u.GithubID
	}
	if u.MicrosoftID != nil {
		user.MicrosoftID = *u.MicrosoftID
	}
	for _, m := range u.AuthMethods {
		user.AuthMethods = append(user.AuthMethods, models.AuthMethod(m))
	}
	if u.MfaEnabled {
		user.TOTPEnabled = u.MfaEnabled
	}
	if u.MfaSecret != nil {
		user.TOTPSecret = *u.MfaSecret
	}
	user.ThemePreference = u.ThemePreference
	if u.LastLoginAt != nil {
		user.LastLoginAt = u.LastLoginAt
	}
	return user
}

// ToTenant converts a sqlc-generated gen.Tenant to a models.Tenant.
func ToTenant(t gen.Tenant) models.Tenant {
	tenant := models.Tenant{
		ID:            t.ID,
		Name:          t.Name,
		BillingStatus: models.BillingStatus(t.BillingStatus),
		IsRoot:        t.IsRoot,
		IsActive:      t.IsActive,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
	if t.PlanID != nil {
		tenant.PlanID = t.PlanID
	}
	if t.StripeCustomerID != nil {
		tenant.StripeCustomerID = *t.StripeCustomerID
	}
	if t.StripeSubscriptionID != nil {
		tenant.StripeSubscriptionID = *t.StripeSubscriptionID
	}
	if t.BillingInterval != nil {
		tenant.BillingInterval = *t.BillingInterval
	}
	tenant.BillingWaived = t.BillingWaiver
	return tenant
}

// ToMembership converts a sqlc-generated gen.TenantMembership to a models.TenantMembership.
func ToMembership(m gen.TenantMembership) models.TenantMembership {
	return models.TenantMembership{
		ID:        m.ID,
		TenantID:  m.TenantID,
		UserID:    m.UserID,
		Role:      models.MemberRole(m.Role),
		JoinedAt:  m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// ToPlan converts a sqlc-generated gen.Plan to a models.Plan.
func ToPlan(p gen.Plan) models.Plan {
	plan := models.Plan{
		ID:                   p.ID,
		Name:                 p.Name,
		Description:          p.Description,
		MonthlyPriceCents:    p.MonthlyPriceCents,
		IncludedSeats:        int(p.IncludedSeats),
		MinSeats:             int(p.MinSeats),
		MaxSeats:             int(p.MaxSeats),
		UsageCreditsPerMonth: int64(p.UsageCreditsPerMonth),
	}
	if p.UserLimit != nil {
		plan.UserLimit = int(*p.UserLimit)
	}
	return plan
}

// ToAPIKey converts a sqlc-generated gen.ApiKey to a models.APIKey.
func ToAPIKey(a gen.ApiKey) models.APIKey {
	key := models.APIKey{
		ID:         a.ID,
		Name:       a.Name,
		KeyHash:    a.KeyHash,
		KeyPreview: a.KeyPreview,
		Authority:  models.APIKeyAuthority(a.Authority),
		IsActive:   a.IsActive,
	}
	if a.TenantID != nil {
	}
	if a.CreatedBy != nil {
		key.CreatedBy = *a.CreatedBy
	}
	if a.LastUsedAt != nil {
		key.LastUsedAt = a.LastUsedAt
	}
	return key
}

// ParseUUID parses a UUID string, returning uuid.Nil on error.
func ParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}
