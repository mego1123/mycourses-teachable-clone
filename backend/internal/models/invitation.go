package models

import (
	"github.com/google/uuid"
	"time"

)

type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
)

type Invitation struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId" validate:"required"`
	Email     string             `json:"email" validate:"required,email"`
	Role      MemberRole         `json:"role" validate:"required,valid_role"`
	Token     string             `json:"-" validate:"required"`
	Status    InvitationStatus   `json:"status" validate:"required,valid_invitation_status"`
	InvitedBy uuid.UUID `json:"invitedBy" validate:"required"`
	ExpiresAt time.Time          `json:"expiresAt" validate:"required"`
	CreatedAt time.Time          `json:"createdAt" validate:"required"`
}
