package models

import (
	"github.com/google/uuid"
	"time"

)

type UsageEvent struct {
	ID        uuid.UUID     `json:"id"`
	TenantID  uuid.UUID     `json:"tenantId" validate:"required"`
	UserID    uuid.UUID     `json:"userId" validate:"required"`
	Type      string                 `json:"type" validate:"required,min=1,max=100"`
	Quantity  int                    `json:"quantity" validate:"required,gte=1"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"createdAt" validate:"required"`
}
