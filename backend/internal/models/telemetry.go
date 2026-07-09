package models

import (
	"github.com/google/uuid"
	"time"

)

// TelemetryEvent represents a single telemetry event for product analytics.
type TelemetryEvent struct {
	ID         uuid.UUID     `json:"id"`
	EventName  string                 `json:"eventName"`
	Category   string                 `json:"category"`
	UserID     *uuid.UUID    `json:"userId,omitempty"`
	TenantID   *uuid.UUID    `json:"tenantId,omitempty"`
	SessionID  string                 `json:"sessionId,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}

// Telemetry categories.
const (
	TelemetryCategoryFunnel     = "funnel"
	TelemetryCategoryEngagement = "engagement"
	TelemetryCategoryCustom     = "custom"
)

// Built-in telemetry event names.
const (
	TelemetryPageView             = "page.view"
	TelemetryUserRegistered       = "user.registered"
	TelemetryUserVerified         = "user.verified"
	TelemetryUserLogin            = "user.login"
	TelemetryCheckoutStarted      = "checkout.started"
	TelemetrySubscriptionActivated = "subscription.activated"
	TelemetrySubscriptionCanceled = "subscription.canceled"
	TelemetryPlanChanged          = "plan.changed"
)
