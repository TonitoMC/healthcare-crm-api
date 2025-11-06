//go:generate mockgen -source=service.go -destination=./mocks/service.go -package=mocks

package user

import (
	roleDomain "github.com/tonitomc/healthcare-crm-api/internal/domain/role"
	roleModels "github.com/tonitomc/healthcare-crm-api/internal/domain/role/models"
	userModels "github.com/tonitomc/healthcare-crm-api/internal/domain/user/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// -----------------------------------------------------------------------------
// Service Interface
// -----------------------------------------------------------------------------

// Service defines the business logic for managing users and their roles.
// Authentication (password hashing/comparison) is handled separately in auth/.
type Service interface {
	// User CRUD

	GetAllUsers() ([]userModels.User, error)
	CreateUser(username, email, passwordHash string) error
	GetByID(id int) (*userModels.User, error)
	GetByUsernameOrEmail(identifier string) (*userModels.User, error)
	UpdateUser(u *userModels.User) error
	DeleteUser(id int) error

	// User → Role management
	GetUserRoles(userID int) ([]roleModels.Role, error)
	AddRole(userID, roleID int) error
	RemoveRole(userID, roleID int) error
	ClearRoles(userID int) error
	GetRolesAndPermissions(userID int) ([]roleModels.Role, []roleModels.Permission, error)
}

// -----------------------------------------------------------------------------
// Implementation
// -----------------------------------------------------------------------------

type service struct {
	repo        Repository
	roleService roleDomain.Service
}

// NewService constructs a new User Service.
func NewService(repo Repository, roleService roleDomain.Service) Service {
	return &service{repo: repo, roleService: roleService}
}

// -----------------------------------------------------------------------------
// User CRUD
// -----------------------------------------------------------------------------

func (s *service) GetAllUsers() ([]userModels.User, error) {
	users, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *service) CreateUser(username, email, passwordHash string) error {
	if username == "" || email == "" || passwordHash == "" {
		return appErr.Wrap("UserService.CreateUser", appErr.ErrInvalidInput, nil)
	}

	u := &userModels.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := s.repo.Create(u); err != nil {
		return err // already wrapped at repository level
	}
	return nil
}

func (s *service) GetByID(id int) (*userModels.User, error) {
	if id <= 0 {
		return nil, appErr.Wrap("UserService.GetByID", appErr.ErrInvalidInput, nil)
	}

	u, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err // repo error (e.g. not found) is bubbled up
	}

	return u, nil
}

func (s *service) GetByUsernameOrEmail(identifier string) (*userModels.User, error) {
	if identifier == "" {
		return nil, appErr.Wrap("UserService.GetByUsernameOrEmail", appErr.ErrInvalidInput, nil)
	}

	u, err := s.repo.GetByUsernameOrEmail(identifier)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *service) UpdateUser(u *userModels.User) error {
	if u == nil || u.ID <= 0 || u.Username == "" || u.Email == "" {
		return appErr.Wrap("UserService.UpdateUser", appErr.ErrInvalidInput, nil)
	}

	if err := s.repo.Update(u); err != nil {
		return err
	}
	return nil
}

func (s *service) DeleteUser(id int) error {
	if id <= 0 {
		return appErr.Wrap("UserService.DeleteUser", appErr.ErrInvalidInput, nil)
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}
	return nil
}

// -----------------------------------------------------------------------------
// User → Role management
// -----------------------------------------------------------------------------

func (s *service) GetUserRoles(userID int) ([]roleModels.Role, error) {
	if userID <= 0 {
		return nil, appErr.Wrap("UserService.GetUserRoles", appErr.ErrInvalidInput, nil)
	}

	if _, err := s.repo.GetByID(userID); err != nil {
		return nil, err
	}

	roles, err := s.repo.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (s *service) AddRole(userID, roleID int) error {
	if userID <= 0 || roleID <= 0 {
		return appErr.Wrap("UserService.AddRole", appErr.ErrInvalidInput, nil)
	}

	if _, err := s.repo.GetByID(userID); err != nil {
		return err
	}

	roles, err := s.repo.GetUserRoles(userID)
	if err != nil {
		return err
	}

	for _, r := range roles {
		if r.ID == roleID {
			// domain-level error → bubble to middleware cleanly
			return appErr.NewDomainError(appErr.ErrConflict, "El usuario ya tiene este rol asignado")
		}
	}

	// --- Proceed normally ---
	if err := s.repo.AddRole(userID, roleID); err != nil {
		return err
	}
	return nil
}

func (s *service) RemoveRole(userID, roleID int) error {
	if userID <= 0 || roleID <= 0 {
		return appErr.Wrap("UserService.RemoveRole", appErr.ErrInvalidInput, nil)
	}

	if _, err := s.repo.GetByID(userID); err != nil {
		return err
	}

	if err := s.repo.RemoveRole(userID, roleID); err != nil {
		return err
	}
	return nil
}

func (s *service) ClearRoles(userID int) error {
	if userID <= 0 {
		return appErr.Wrap("UserService.ClearRoles", appErr.ErrInvalidInput, nil)
	}

	if _, err := s.repo.GetByID(userID); err != nil {
		return err
	}

	if err := s.repo.ClearRoles(userID); err != nil {
		return err
	}
	return nil
}

func (s *service) GetRolesAndPermissions(userID int) ([]roleModels.Role, []roleModels.Permission, error) {
	if userID <= 0 {
		return nil, nil, appErr.Wrap("UserService.GetRolesAndPermissions", appErr.ErrInvalidInput, nil)
	}

	if _, err := s.repo.GetByID(userID); err != nil {
		return nil, nil, err
	}

	roles, err := s.repo.GetUserRoles(userID)
	if err != nil {
		return nil, nil, err
	}

	var allPerms []roleModels.Permission
	permSeen := make(map[int]bool)

	for _, r := range roles {
		perms, err := s.roleService.GetPermissions(r.ID)
		if err != nil {
			// Bubble up, but contextualize which role failed
			return nil, nil, appErr.Wrap(
				"UserService.GetRolesAndPermissions(role fetch)",
				appErr.ErrInternal,
				err,
			)
		}

		for _, p := range perms {
			if !permSeen[p.ID] {
				allPerms = append(allPerms, p)
				permSeen[p.ID] = true
			}
		}
	}

	return roles, allPerms, nil
}
