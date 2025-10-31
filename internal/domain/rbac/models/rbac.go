// internal/domain/rbac/models/rbac.go
package models

import (
	roleModels "github.com/tonitomc/healthcare-crm-api/internal/domain/role/models"
	userModels "github.com/tonitomc/healthcare-crm-api/internal/domain/user/models"
)

// RBAC represents the complete access context for a user.
// It is *not* persisted in the database — it’s dynamically built
// by joining users, roles, and permissions across domains.

// This model is primarily used by the Auth layer to build JWT claims.
type RBAC struct {
	User        *userModels.User        `json:"user"`
	Roles       []roleModels.Role       `json:"roles"`
	Permissions []roleModels.Permission `json:"permissions"`
}
