// internal/domain/rbac/service.go
package rbac

import (
	"github.com/tonitomc/healthcare-crm-api/internal/domain/rbac/models"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/role"
	roleModels "github.com/tonitomc/healthcare-crm-api/internal/domain/role/models"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/user"
	userModels "github.com/tonitomc/healthcare-crm-api/internal/domain/user/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// -----------------------------------------------------------------------------
// Service Interface
// -----------------------------------------------------------------------------

// Service provides unified access resolution (user → roles → permissions).
// It aggregates data from both the User and Role domains to construct
// a complete RBAC access context, used mainly by the Auth layer.
type Service interface {
	GetUserAccess(userID int) (*models.RBAC, error)
}

// -----------------------------------------------------------------------------
// Implementation
// -----------------------------------------------------------------------------

type service struct {
	userService user.Service
	roleService role.Service
}

// NewService constructs a new RBAC service.
// It requires both User and Role services to resolve the access graph.
func NewService(userService user.Service, roleService role.Service) Service {
	return &service{
		userService: userService,
		roleService: roleService,
	}
}

// -----------------------------------------------------------------------------
// RBAC Resolution
// -----------------------------------------------------------------------------

// GetUserAccess resolves a full RBAC context (User, Roles, Permissions) for the given user.
func (s *service) GetUserAccess(userID int) (*models.RBAC, error) {
	if userID <= 0 {
		return nil, appErr.Wrap("RBACService.GetUserAccess", appErr.ErrInvalidInput, nil)
	}

	userData, err := s.userService.GetByID(userID)
	if err != nil {
		return nil, appErr.Wrap("RBACService.GetUserAccess(user)", appErr.ErrNotFound, err)
	}

	roles, perms, err := s.userService.GetRolesAndPermissions(userID)
	if err != nil {
		return nil, appErr.Wrap("RBACService.GetUserAccess(roles+perms)", appErr.ErrInternal, err)
	}

	rbacCtx := &models.RBAC{
		User:        &userModels.User{ID: userData.ID, Username: userData.Username, Email: userData.Email},
		Roles:       make([]roleModels.Role, len(roles)),
		Permissions: make([]roleModels.Permission, len(perms)),
	}

	copy(rbacCtx.Roles, roles)
	copy(rbacCtx.Permissions, perms)

	return rbacCtx, nil
}
