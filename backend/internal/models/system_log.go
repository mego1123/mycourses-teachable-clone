package models

import (
	"github.com/google/uuid"
	"time"

)

type LogSeverity string

const (
	LogCritical LogSeverity = "critical"
	LogHigh     LogSeverity = "high"
	LogMedium   LogSeverity = "medium"
	LogLow      LogSeverity = "low"
	LogDebug    LogSeverity = "debug"
)

type LogCategory string

const (
	LogCatAuth     LogCategory = "auth"
	LogCatBilling  LogCategory = "billing"
	LogCatAdmin    LogCategory = "admin"
	LogCatSystem   LogCategory = "system"
	LogCatSecurity LogCategory = "security"
	LogCatTenant   LogCategory = "tenant"
)

type SystemLog struct {
	ID        uuid.UUID     `json:"id"`
	Severity  LogSeverity            `json:"severity"`
	Category  LogCategory            `json:"category,omitempty"`
	Message   string                 `json:"message"`
	UserID    *uuid.UUID    `json:"userId,omitempty"`
	TenantID  *uuid.UUID    `json:"tenantId,omitempty"`
	Action    string                 `json:"action,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"createdAt"`
}
