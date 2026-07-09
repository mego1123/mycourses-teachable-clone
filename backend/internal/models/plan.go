package models

import (
	"github.com/google/uuid"
	"time"

)

type EntitlementType string

const (
	EntitlementTypeBool    EntitlementType = "bool"
	EntitlementTypeNumeric EntitlementType = "numeric"
)

type CreditResetPolicy string

const (
	CreditResetPolicyReset  CreditResetPolicy = "reset"
	CreditResetPolicyAccrue CreditResetPolicy = "accrue"
)

type EntitlementValue struct {
	Type         EntitlementType `json:"type"`
	BoolValue    bool            `json:"boolValue"`
	NumericValue int64           `json:"numericValue"`
	Description  string          `json:"description"`
}

type PricingModel string

const (
	PricingModelFlat    PricingModel = "flat"
	PricingModelPerSeat PricingModel = "per_seat"
)

type Plan struct {
	ID                   uuid.UUID          `json:"id"`
	Name                 string                      `json:"name" validate:"required,min=1,max=200"`
	Description          string                      `json:"description"`
	PricingModel         PricingModel                `json:"pricingModel" validate:"required,valid_pricing_model"`
	MonthlyPriceCents    int64                       `json:"monthlyPriceCents" validate:"gte=0"`
	AnnualDiscountPct    int                         `json:"annualDiscountPct" validate:"gte=0,lte=100"`
	PerSeatPriceCents    int64                       `json:"perSeatPriceCents" validate:"gte=0"`
	IncludedSeats        int                         `json:"includedSeats" validate:"gte=0"`
	MinSeats             int                         `json:"minSeats" validate:"gte=0"`
	MaxSeats             int                         `json:"maxSeats" validate:"gte=0"`
	UsageCreditsPerMonth int64                       `json:"usageCreditsPerMonth" validate:"gte=0"`
	CreditResetPolicy    CreditResetPolicy           `json:"creditResetPolicy" validate:"required,valid_credit_reset"`
	BonusCredits         int64                       `json:"bonusCredits" validate:"gte=0"`
	UserLimit            int                         `json:"userLimit" validate:"gte=0"`
	TrialDays            int                         `json:"trialDays" validate:"gte=0"`
	Entitlements         map[string]EntitlementValue `json:"entitlements"`
	IsSystem             bool                        `json:"isSystem"`
	IsArchived           bool                        `json:"isArchived"`
	CreatedAt            time.Time                   `json:"createdAt" validate:"required"`
	UpdatedAt            time.Time                   `json:"updatedAt" validate:"required"`
}
