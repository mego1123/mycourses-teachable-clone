package models

import (
	"github.com/google/uuid"
	"time"

)

type ConfigVarType string

const (
	ConfigTypeString   ConfigVarType = "string"
	ConfigTypeNumeric  ConfigVarType = "numeric"
	ConfigTypeEnum     ConfigVarType = "enum"
	ConfigTypeTemplate ConfigVarType = "template"
)

type ConfigVar struct {
	ID          uuid.UUID `json:"id"`
	Name        string             `json:"name" validate:"required,min=1,max=200"`
	Description string             `json:"description"`
	Type        ConfigVarType      `json:"type" validate:"required,valid_config_type"`
	Value       string             `json:"value"`
	Options     string             `json:"options,omitempty"`
	IsSystem    bool               `json:"isSystem"`
	CreatedAt   time.Time          `json:"createdAt" validate:"required"`
	UpdatedAt   time.Time          `json:"updatedAt" validate:"required"`
}

func ValidConfigVarType(t ConfigVarType) bool {
	switch t {
	case ConfigTypeString, ConfigTypeNumeric, ConfigTypeEnum, ConfigTypeTemplate:
		return true
	}
	return false
}
