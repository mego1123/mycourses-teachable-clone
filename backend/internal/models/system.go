package models

import (
	"github.com/google/uuid"
	"time"

)

type SystemConfig struct {
	ID            uuid.UUID  `json:"id"`
	Initialized   bool                `json:"initialized"`
	InitializedAt *time.Time          `json:"initializedAt,omitempty"`
	InitializedBy *uuid.UUID `json:"-"`
	Version       string              `json:"version"`
}
