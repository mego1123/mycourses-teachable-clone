package models

import (
	"github.com/google/uuid"
	"time"

)

type SSOConnection struct {
	ID               uuid.UUID `json:"id"`
	TenantID         uuid.UUID `json:"tenantId" validate:"required"`
	IdPMetadataURL   string             `json:"idpMetadataUrl"`
	IdPMetadataXML   string             `json:"-"`
	IdPEntityID      string             `json:"idpEntityId" validate:"required"`
	IdPSSOURL        string             `json:"idpSsoUrl" validate:"required,url"`
	IdPCertificate   string             `json:"-" validate:"required"`
	AttributeMapping SSOAttributeMap    `json:"attributeMapping"`
	IsActive         bool               `json:"isActive"`
	CreatedAt        time.Time          `json:"createdAt" validate:"required"`
	UpdatedAt        time.Time          `json:"updatedAt" validate:"required"`
}

type SSOAttributeMap struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
}
