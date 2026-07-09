package models

import (
	"github.com/google/uuid"
	"time"

)

type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Name      string             `json:"name" validate:"required,min=1,max=200"`
	Slug      string             `json:"slug" validate:"required,min=1,max=100"`
	IsRoot    bool               `json:"isRoot"`
	IsActive  bool                `json:"isActive"`
	PlanID               *uuid.UUID `json:"planId,omitempty"`
	BillingWaived        bool               `json:"billingWaived"`
	SubscriptionCredits  int64              `json:"subscriptionCredits"`
	PurchasedCredits     int64              `json:"purchasedCredits"`
	StripeCustomerID     string             `json:"stripeCustomerId,omitempty"`
	BillingStatus        BillingStatus      `json:"billingStatus" validate:"omitempty,valid_billing_status"`
	StripeSubscriptionID string             `json:"stripeSubscriptionId,omitempty"`
	BillingInterval      string             `json:"billingInterval,omitempty"`
	CurrentPeriodEnd     *time.Time         `json:"currentPeriodEnd,omitempty"`
	CanceledAt           *time.Time         `json:"canceledAt,omitempty"`
	TrialUsedAt          *time.Time         `json:"trialUsedAt,omitempty"`
	SeatQuantity         int                `json:"seatQuantity"`
	CreatedAt            time.Time          `json:"createdAt" validate:"required"`
	UpdatedAt            time.Time          `json:"updatedAt" validate:"required"`
}
