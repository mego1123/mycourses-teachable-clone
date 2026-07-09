package models

import (
	"github.com/google/uuid"
	"time"

)

// BillingStatus represents the current billing state of a tenant.
type BillingStatus string

const (
	BillingStatusNone     BillingStatus = "none"
	BillingStatusActive   BillingStatus = "active"
	BillingStatusPastDue  BillingStatus = "past_due"
	BillingStatusCanceled BillingStatus = "canceled"
)

// TransactionType categorizes financial transactions.
type TransactionType string

const (
	TransactionSubscription   TransactionType = "subscription"
	TransactionCreditPurchase TransactionType = "credit_purchase"
	TransactionRefund         TransactionType = "refund"
)

// FinancialTransaction records every payment event.
type FinancialTransaction struct {
	ID                   uuid.UUID  `json:"id"`
	TenantID             uuid.UUID  `json:"tenantId" validate:"required"`
	UserID               uuid.UUID  `json:"userId" validate:"required"`
	Type                 TransactionType     `json:"type" validate:"required,oneof=subscription credit_purchase refund"`
	AmountCents          int64               `json:"amountCents"`
	SubtotalCents        int64               `json:"subtotalCents"`
	TaxAmountCents       int64               `json:"taxAmountCents"`
	Currency             string              `json:"currency" validate:"required,len=3"`
	Description          string              `json:"description"`
	InvoiceNumber        string              `json:"invoiceNumber" validate:"required"`
	StripeSessionID      string              `json:"stripeSessionId,omitempty"`
	StripeInvoiceID      string              `json:"stripeInvoiceId,omitempty"`
	StripeSubscriptionID string              `json:"stripeSubscriptionId,omitempty"`
	PlanID               *uuid.UUID `json:"planId,omitempty"`
	PlanName             string              `json:"planName,omitempty"`
	BundleID             *uuid.UUID `json:"bundleId,omitempty"`
	BundleName           string              `json:"bundleName,omitempty"`
	BillingInterval      string              `json:"billingInterval,omitempty"`
	CreatedAt            time.Time           `json:"createdAt" validate:"required"`
}

// StripeMapping maps internal entities (plans, bundles) to Stripe Products/Prices.
type StripeMapping struct {
	ID              uuid.UUID 
	EntityType      string             
	EntityID        uuid.UUID 
	StripePriceID   string             
	StripeProductID string             
	CreatedAt       time.Time          
}

// InvoiceCounter is used for atomic invoice number generation.
type InvoiceCounter struct {
	ID    string 
	Value int64  
}

// DailyMetric stores daily business metrics for dashboard charts.
type DailyMetric struct {
	ID        uuid.UUID 
	Date      string             `json:"date"`
	DAU       int64              `json:"dau"`
	WAU       int64              `json:"wau"`
	MAU       int64              `json:"mau"`
	Revenue   int64              `json:"revenue"`
	ARR       int64              `json:"arr"`
	CreatedAt time.Time          `json:"createdAt"`
}
