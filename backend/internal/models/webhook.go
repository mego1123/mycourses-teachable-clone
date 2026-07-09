package models

import (
	"github.com/google/uuid"
	"time"

)

type WebhookEventType string

const (
	// Billing events (Tier 1)
	WebhookEventSubscriptionActivated WebhookEventType = "subscription.activated"
	WebhookEventSubscriptionCanceled  WebhookEventType = "subscription.canceled"
	WebhookEventPaymentReceived       WebhookEventType = "payment.received"
	WebhookEventPaymentFailed         WebhookEventType = "payment.failed"

	// Team lifecycle events (Tier 2)
	WebhookEventMemberInvited        WebhookEventType = "member.invited"
	WebhookEventMemberJoined         WebhookEventType = "member.joined"
	WebhookEventMemberRemoved        WebhookEventType = "member.removed"
	WebhookEventMemberRoleChanged    WebhookEventType = "member.role_changed"
	WebhookEventOwnershipTransferred WebhookEventType = "ownership.transferred"

	// User lifecycle events (Tier 3)
	WebhookEventUserRegistered  WebhookEventType = "user.registered"
	WebhookEventUserVerified    WebhookEventType = "user.verified"
	WebhookEventUserDeactivated WebhookEventType = "user.deactivated"

	// Credits & billing details (Tier 4)
	WebhookEventCreditsPurchased  WebhookEventType = "credits.purchased"
	WebhookEventPlanChanged       WebhookEventType = "plan.changed"
	WebhookEventTenantCreated     WebhookEventType = "tenant.created"
	WebhookEventTenantDeactivated WebhookEventType = "tenant.deactivated"

	// Audit & security events (Tier 5)
	WebhookEventUserDeleted    WebhookEventType = "user.deleted"
	WebhookEventTenantDeleted  WebhookEventType = "tenant.deleted"
	WebhookEventAPIKeyCreated  WebhookEventType = "api_key.created"
	WebhookEventAPIKeyRevoked  WebhookEventType = "api_key.revoked"
)

// AllWebhookEventTypes lists every supported webhook event type.
var AllWebhookEventTypes = []WebhookEventType{
	// Tier 1: Billing
	WebhookEventSubscriptionActivated,
	WebhookEventSubscriptionCanceled,
	WebhookEventPaymentReceived,
	WebhookEventPaymentFailed,
	// Tier 2: Team lifecycle
	WebhookEventMemberInvited,
	WebhookEventMemberJoined,
	WebhookEventMemberRemoved,
	WebhookEventMemberRoleChanged,
	WebhookEventOwnershipTransferred,
	// Tier 3: User lifecycle
	WebhookEventUserRegistered,
	WebhookEventUserVerified,
	WebhookEventUserDeactivated,
	// Tier 4: Credits & billing details
	WebhookEventCreditsPurchased,
	WebhookEventPlanChanged,
	WebhookEventTenantCreated,
	WebhookEventTenantDeactivated,
	// Tier 5: Audit & security
	WebhookEventUserDeleted,
	WebhookEventTenantDeleted,
	WebhookEventAPIKeyCreated,
	WebhookEventAPIKeyRevoked,
}

func ValidWebhookEventType(e WebhookEventType) bool {
	for _, t := range AllWebhookEventTypes {
		if t == e {
			return true
		}
	}
	return false
}

type Webhook struct {
	ID            uuid.UUID `json:"id"`
	Name          string             `json:"name" validate:"required,min=1,max=100"`
	Description   string             `json:"description"`
	URL           string             `json:"url" validate:"required,url"`
	Secret        string             `json:"-" validate:"required"`
	SecretPreview string             `json:"secretPreview" validate:"required"`
	Events        []WebhookEventType `json:"events" validate:"required,min=1,dive,valid_webhook_event"`
	IsActive      bool               `json:"isActive"`
	CreatedBy     uuid.UUID `json:"createdBy" validate:"required"`
	CreatedAt     time.Time          `json:"createdAt" validate:"required"`
	UpdatedAt     time.Time          `json:"updatedAt" validate:"required"`
}

type WebhookDelivery struct {
	ID           uuid.UUID `json:"id"`
	WebhookID    uuid.UUID `json:"webhookId"`
	EventType    WebhookEventType   `json:"eventType"`
	Payload      string             `json:"payload"`
	ResponseCode int                `json:"responseCode"`
	ResponseBody string             `json:"responseBody"`
	Success      bool               `json:"success"`
	Duration     int64              `json:"durationMs"`
	RetryCount   int                `json:"retryCount"`
	MaxRetries   int                `json:"maxRetries"`
	CreatedAt    time.Time          `json:"createdAt"`
}
