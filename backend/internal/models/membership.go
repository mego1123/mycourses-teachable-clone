package models

import (
	"github.com/google/uuid"
	"time"

)

type MemberRole string

const (
	RoleOwner MemberRole = "owner"
	RoleAdmin MemberRole = "admin"
	RoleUser  MemberRole = "user"
)

type TenantMembership struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId" validate:"required"`
	TenantID  uuid.UUID `json:"tenantId" validate:"required"`
	Role      MemberRole         `json:"role" validate:"required,valid_role"`
	JoinedAt  time.Time          `json:"joinedAt" validate:"required"`
	UpdatedAt time.Time          `json:"updatedAt" validate:"required"`
}

var roleHierarchy = map[MemberRole]int{
	RoleUser:  1,
	RoleAdmin: 2,
	RoleOwner: 3,
}

func RoleHasPermission(userRole MemberRole, requiredRole MemberRole) bool {
	return roleHierarchy[userRole] >= roleHierarchy[requiredRole]
}

func ValidRole(role MemberRole) bool {
	_, ok := roleHierarchy[role]
	return ok
}
