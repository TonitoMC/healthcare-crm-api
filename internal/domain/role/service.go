package role

import (
	"github.com/tonitomc/healthcare-crm-api/internal/domain/role/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Service interface {
	GetAllRoles() ([]models.Role, error)
	GetRoleByID(id int) (*models.Role, []models.Permission, error)
	CreateRole(role *models.Role) error
	UpdateRole(role *models.Role) error
	DeleteRole(id int) error

	GetPermissions(roleID int) ([]models.Permission, error)
	UpdateRolePermissions(roleID int, permissionIDs []int) error
	AddPermission(roleID, permissionID int) error
	RemovePermission(roleID, permissionID int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// -----------------------------------------------------------------------------
// Role CRUD
// -----------------------------------------------------------------------------

func (s *service) GetAllRoles() ([]models.Role, error) {
	roles, err := s.repo.GetAll()
	if err != nil {
		return nil, err // repo already wrapped
	}
	return roles, nil
}

func (s *service) GetRoleByID(id int) (*models.Role, []models.Permission, error) {
	if id <= 0 {
		return nil, nil, appErr.Wrap("roleService.GetRoleByID", appErr.ErrInvalidInput, nil)
	}

	role, err := s.repo.GetByID(id)
	if err != nil {
		return nil, nil, err
	}

	perms, err := s.repo.GetPermissions(id)
	if err != nil {
		return role, nil, err
	}

	return role, perms, nil
}

func (s *service) CreateRole(role *models.Role) error {
	if role == nil || role.Name == "" || role.Description == "" {
		return appErr.Wrap("roleService.CreateRole", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Create(role)
}

func (s *service) UpdateRole(role *models.Role) error {
	if role == nil || role.ID <= 0 {
		return appErr.Wrap("roleService.UpdateRole", appErr.ErrInvalidInput, nil)
	}

	if err := s.repo.Update(role); err != nil {
		return err
	}
	return nil
}

func (s *service) DeleteRole(id int) error {
	if id <= 0 {
		return appErr.Wrap("roleService.DeleteRole", appErr.ErrInvalidInput, nil)
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}
	return nil
}

// -----------------------------------------------------------------------------
// Permissions
// -----------------------------------------------------------------------------

func (s *service) GetPermissions(roleID int) ([]models.Permission, error) {
	if roleID <= 0 {
		return nil, appErr.Wrap("roleService.GetPermissions", appErr.ErrInvalidInput, nil)
	}
	perms, err := s.repo.GetPermissions(roleID)
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func (s *service) AddPermission(roleID, permissionID int) error {
	if roleID <= 0 || permissionID <= 0 {
		return appErr.Wrap("roleService.AddPermission", appErr.ErrInvalidInput, nil)
	}
	if err := s.repo.AddPermission(roleID, permissionID); err != nil {
		return err
	}
	return nil
}

func (s *service) RemovePermission(roleID, permissionID int) error {
	if roleID <= 0 || permissionID <= 0 {
		return appErr.Wrap("roleService.RemovePermission", appErr.ErrInvalidInput, nil)
	}
	if err := s.repo.RemovePermission(roleID, permissionID); err != nil {
		return err
	}
	return nil
}

func (s *service) UpdateRolePermissions(roleID int, permissionIDs []int) error {
	if roleID <= 0 {
		return appErr.Wrap("roleService.UpdateRolePermissions", appErr.ErrInvalidInput, nil)
	}

	// Clear existing ones first
	if err := s.repo.ClearPermissions(roleID); err != nil {
		return err
	}

	// Add back
	for _, pid := range permissionIDs {
		if pid <= 0 {
			continue
		}
		if err := s.repo.AddPermission(roleID, pid); err != nil {
			return appErr.Wrap("roleService.UpdateRolePermissions", appErr.ErrConflict, err)
		}
	}

	return nil
}
