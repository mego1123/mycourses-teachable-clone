package models

import (
	"github.com/google/uuid"
	"time"

)

// EventDefinition represents a user-defined event type with optional parent dependency.
type EventDefinition struct {
	ID          uuid.UUID  `json:"id"`
	Name        string              `json:"name" validate:"required,min=1,max=128"`
	Description string              `json:"description" validate:"max=256"`
	ParentID    *uuid.UUID `json:"parentId,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}
