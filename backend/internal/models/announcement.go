package models

import (
	"github.com/google/uuid"
	"time"

)

type Announcement struct {
	ID          uuid.UUID `json:"id"`
	Title       string             `json:"title" validate:"required,min=1,max=200"`
	Body        string             `json:"body" validate:"required,min=1"`
	IsPublished bool               `json:"isPublished"`
	PublishedAt *time.Time         `json:"publishedAt,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" validate:"required"`
	UpdatedAt   time.Time          `json:"updatedAt" validate:"required"`
}
