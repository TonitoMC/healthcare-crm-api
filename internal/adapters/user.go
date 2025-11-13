package adapters

import (
	middlewarePkg "github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	roleModels "github.com/tonitomc/healthcare-crm-api/internal/domain/role/models"
	userDomain "github.com/tonitomc/healthcare-crm-api/internal/domain/user"
)

// UserPermissionAdapter adapts user.Service to the middleware's PermissionProvider interface.
type UserPermissionAdapter struct {
	Service userDomain.Service
}

func NewUserPermissionAdapter(service userDomain.Service) *UserPermissionAdapter {
	return &UserPermissionAdapter{Service: service}
}

// Implements middleware.PermissionProvider
func (u *UserPermissionAdapter) GetRolesAndPermissions(userID int) ([]any, []middlewarePkg.PermissionLike, error) {
	_, perms, err := u.Service.GetRolesAndPermissions(userID)
	if err != nil {
		return nil, nil, err
	}

	out := make([]middlewarePkg.PermissionLike, len(perms))
	for i := range perms {
		out[i] = rolePermissionWrapper{perm: perms[i]}
	}
	return nil, out, nil
}

// rolePermissionWrapper implements middleware.PermissionLike
type rolePermissionWrapper struct {
	perm roleModels.Permission
}

func (r rolePermissionWrapper) GetName() string {
	return r.perm.Name
}
