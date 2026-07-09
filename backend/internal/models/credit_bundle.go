package models

import (
	"github.com/google/uuid"
	"time"

)

type CreditBundle struct {
	ID         uuid.UUID `json:"id"`
	Name       string             `json:"name" validate:"required,min=1,max=200"`
	Credits    int64              `json:"credits" validate:"required,gt=0"`
	PriceCents int64              `json:"priceCents" validate:"required,gt=0"`
	IsActive   bool               `json:"isActive"`
	SortOrder  int                `json:"sortOrder"`
	CreatedAt  time.Time          `json:"createdAt" validate:"required"`
	UpdatedAt  time.Time          `json:"updatedAt" validate:"required"`
}
