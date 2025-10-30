package user

import (
	"fmt"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/user/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// This service is in charge of all the CRUD operations related to Users, it's
// important to note password hashing & comparisons do NOT happen here & are instead
// in the auth/ module.

// Service holds all the dependencies
type Service struct {
	repo *Repository
}

// NewService creates a new Service instance
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateUser hashes the password and stores a new user
func (s *Service) CreateUser(username, email, passwordHash string) error {
	if username == "" || email == "" || passwordHash == "" {
		return fmt.Errorf("CreateUser: %w", appErr.ErrInvalidInput)
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := s.repo.Create(user); err != nil {
		return fmt.Errorf("CreateUser: %w", err)
	}

	return nil
}

// GetByID fetches a user by ID
func (s *Service) GetByID(id int) (*models.User, error) {
	u, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return u, nil
}

// GetByUsernameOrEmail returns a user by either username or email
func (s *Service) GetByUsernameOrEmail(identifier string) (*models.User, error) {
	u, err := s.repo.GetByUsernameOrEmail(identifier)
	if err != nil {
		return nil, fmt.Errorf("GetByUsernameOrEmail: %w", err)
	}
	return u, nil
}

// DeleteUser deletes a user by ID
func (s *Service) DeleteUser(id int) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("DeleteUser: %w", err)
	}
	return nil
}

func (s *Service) GetRolesAndPermissions(userID int) ([]models.Role, []models.Permission, error) {
	roles, perms, err := s.repo.GetRolesAndPermissions(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("GetRolesAndPermissions: %w", err)
	}
	return roles, perms, nil
}
