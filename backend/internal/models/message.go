package models

import (
	"github.com/google/uuid"
	"time"

)

type Message struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId" validate:"required"`
	Subject   string             `json:"subject" validate:"required,min=1,max=200"`
	Body      string             `json:"body" validate:"required,min=1"`
	IsSystem  bool               `json:"isSystem"`
	Read      bool               `json:"read"`
	CreatedAt time.Time          `json:"createdAt" validate:"required"`
}
