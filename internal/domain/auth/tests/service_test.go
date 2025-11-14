package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/auth"
	rbacModels "github.com/tonitomc/healthcare-crm-api/internal/domain/rbac/models"
	roleModels "github.com/tonitomc/healthcare-crm-api/internal/domain/role/models"
	userModels "github.com/tonitomc/healthcare-crm-api/internal/domain/user/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

const testSecret = "test-secret-key-for-jwt"

func TestRegister(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		mockUser := &mockUserService{}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
			AccessTTL: 1 * time.Hour,
			Issuer:    "test-issuer",
		})

		err := svc.Register("testuser", "test@example.com", "password123")
		require.NoError(t, err)
		require.Equal(t, "testuser", mockUser.lastCreatedUsername)
	})

	t.Run("Empty Username", func(t *testing.T) {
		mockUser := &mockUserService{}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		err := svc.Register("", "test@example.com", "password123")
		require.Error(t, err)
	})

	t.Run("Empty Email", func(t *testing.T) {
		mockUser := &mockUserService{}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		err := svc.Register("testuser", "", "password123")
		require.Error(t, err)
	})

	t.Run("Empty Password", func(t *testing.T) {
		mockUser := &mockUserService{}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		err := svc.Register("testuser", "test@example.com", "")
		require.Error(t, err)
	})

	t.Run("User Creation Fails", func(t *testing.T) {
		mockUser := &mockUserService{createError: true}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		err := svc.Register("testuser", "test@example.com", "password123")
		require.Error(t, err)
	})
}

func TestLogin(t *testing.T) {
	t.Parallel()

	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	t.Run("Success", func(t *testing.T) {
		mockUser := &mockUserService{
			user: &userModels.User{
				ID:           1,
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: string(hashedPass),
			},
		}
		mockRBAC := &mockRBACService{
			rbac: &rbacModels.RBAC{
				User: &userModels.User{
					ID:       1,
					Username: "testuser",
				},
				Roles: []roleModels.Role{
					{ID: 1, Name: "admin"},
				},
				Permissions: []roleModels.Permission{
					{ID: 1, Name: "manage-users"},
				},
			},
		}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
			AccessTTL: 1 * time.Hour,
			Issuer:    "test-issuer",
		})

		token, err := svc.Login("testuser", "password123")
		require.NoError(t, err)
		require.NotEmpty(t, token)
	})

	t.Run("Empty Identifier", func(t *testing.T) {
		mockUser := &mockUserService{}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		_, err := svc.Login("", "password123")
		require.Error(t, err)
	})

	t.Run("Empty Password", func(t *testing.T) {
		mockUser := &mockUserService{}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		_, err := svc.Login("testuser", "")
		require.Error(t, err)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockUser := &mockUserService{getUserError: true}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		_, err := svc.Login("nonexistent", "password123")
		require.Error(t, err)
	})

	t.Run("Wrong Password", func(t *testing.T) {
		mockUser := &mockUserService{
			user: &userModels.User{
				ID:           1,
				Username:     "testuser",
				PasswordHash: string(hashedPass),
			},
		}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		_, err := svc.Login("testuser", "wrongpassword")
		require.Error(t, err)
	})

	t.Run("RBAC Service Fails", func(t *testing.T) {
		mockUser := &mockUserService{
			user: &userModels.User{
				ID:           1,
				Username:     "testuser",
				PasswordHash: string(hashedPass),
			},
		}
		mockRBAC := &mockRBACService{rbacError: true}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		_, err := svc.Login("testuser", "password123")
		require.Error(t, err)
	})
}

func TestValidateToken(t *testing.T) {
	t.Parallel()

	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	mockUser := &mockUserService{
		user: &userModels.User{
			ID:           1,
			Username:     "testuser",
			PasswordHash: string(hashedPass),
		},
	}
	mockRBAC := &mockRBACService{
		rbac: &rbacModels.RBAC{
			User: &userModels.User{
				ID:       1,
				Username: "testuser",
			},
			Roles:       []roleModels.Role{{ID: 1, Name: "admin"}},
			Permissions: []roleModels.Permission{{ID: 1, Name: "manage-users"}},
		},
	}

	svc := auth.NewService(mockUser, mockRBAC, auth.Config{
		JWTSecret: testSecret,
		AccessTTL: 1 * time.Hour,
		Issuer:    "test-issuer",
	})

	// First login to get a valid token
	token, err := svc.Login("testuser", "password123")
	require.NoError(t, err)

	t.Run("Valid Token", func(t *testing.T) {
		parsedToken, claims, err := svc.ValidateToken(token)
		require.NoError(t, err)
		require.NotNil(t, parsedToken)
		require.NotNil(t, claims)
		require.Equal(t, 1, claims.UserID)
		require.Equal(t, "testuser", claims.Username)
		require.Contains(t, claims.Roles, "admin")
		require.Contains(t, claims.Permissions, "manage-users")
	})

	t.Run("Empty Token", func(t *testing.T) {
		_, _, err := svc.ValidateToken("")
		require.Error(t, err)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		_, _, err := svc.ValidateToken("invalid.token.string")
		require.Error(t, err)
	})

	t.Run("Wrong Issuer", func(t *testing.T) {
		// Create service with different issuer
		svc2 := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
			AccessTTL: 1 * time.Hour,
			Issuer:    "different-issuer",
		})

		_, _, err := svc2.ValidateToken(token)
		require.Error(t, err)
	})

	t.Run("Wrong Secret", func(t *testing.T) {
		// Create service with different secret
		svc3 := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: "different-secret",
			AccessTTL: 1 * time.Hour,
			Issuer:    "test-issuer",
		})

		_, _, err := svc3.ValidateToken(token)
		require.Error(t, err)
	})
}

func TestChangePassword(t *testing.T) {
	t.Parallel()

	hashedOldPass, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)

	t.Run("Success", func(t *testing.T) {
		mockUser := &mockUserService{
			user: &userModels.User{
				ID:           1,
				Username:     "testuser",
				PasswordHash: string(hashedOldPass),
			},
		}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		err := svc.ChangePassword(1, "oldpassword", "newpassword123")
		require.NoError(t, err)
		require.True(t, mockUser.updateCalled)
	})

	t.Run("Invalid UserID", func(t *testing.T) {
		mockUser := &mockUserService{}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		err := svc.ChangePassword(0, "oldpassword", "newpassword123")
		require.Error(t, err)
	})

	t.Run("Empty Old Password", func(t *testing.T) {
		mockUser := &mockUserService{}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		err := svc.ChangePassword(1, "", "newpassword123")
		require.Error(t, err)
	})

	t.Run("Empty New Password", func(t *testing.T) {
		mockUser := &mockUserService{}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		err := svc.ChangePassword(1, "oldpassword", "")
		require.Error(t, err)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockUser := &mockUserService{getUserError: true}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		err := svc.ChangePassword(1, "oldpassword", "newpassword123")
		require.Error(t, err)
	})

	t.Run("Wrong Old Password", func(t *testing.T) {
		mockUser := &mockUserService{
			user: &userModels.User{
				ID:           1,
				Username:     "testuser",
				PasswordHash: string(hashedOldPass),
			},
		}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		err := svc.ChangePassword(1, "wrongpassword", "newpassword123")
		require.Error(t, err)
	})

	t.Run("Update User Fails", func(t *testing.T) {
		mockUser := &mockUserService{
			user: &userModels.User{
				ID:           1,
				Username:     "testuser",
				PasswordHash: string(hashedOldPass),
			},
			updateError: true,
		}
		mockRBAC := &mockRBACService{}

		svc := auth.NewService(mockUser, mockRBAC, auth.Config{
			JWTSecret: testSecret,
		})

		err := svc.ChangePassword(1, "oldpassword", "newpassword123")
		require.Error(t, err)
	})
}

// Mock implementations
type mockUserService struct {
	user                *userModels.User
	createError         bool
	getUserError        bool
	updateError         bool
	lastCreatedUsername string
	updateCalled        bool
}

func (m *mockUserService) CreateUser(username, email, passwordHash string) error {
	if m.createError {
		return appErr.ErrInternal
	}
	m.lastCreatedUsername = username
	return nil
}

func (m *mockUserService) GetByUsernameOrEmail(identifier string) (*userModels.User, error) {
	if m.getUserError {
		return nil, appErr.ErrNotFound
	}
	if m.user == nil {
		return nil, appErr.ErrNotFound
	}
	return m.user, nil
}

func (m *mockUserService) GetByID(id int) (*userModels.User, error) {
	if m.getUserError {
		return nil, appErr.ErrNotFound
	}
	if m.user == nil {
		return nil, appErr.ErrNotFound
	}
	return m.user, nil
}

func (m *mockUserService) UpdateUser(user *userModels.User) error {
	if m.updateError {
		return appErr.ErrInternal
	}
	m.updateCalled = true
	return nil
}

func (m *mockUserService) GetAllUsers() ([]userModels.User, error) {
	return nil, nil
}

func (m *mockUserService) DeleteUser(id int) error {
	return nil
}

func (m *mockUserService) AddRole(userID, roleID int) error {
	return nil
}

func (m *mockUserService) RemoveRole(userID, roleID int) error {
	return nil
}

func (m *mockUserService) GetUserRoles(userID int) ([]roleModels.Role, error) {
	return nil, nil
}

func (m *mockUserService) ClearRoles(userID int) error {
	return nil
}

func (m *mockUserService) GetRolesAndPermissions(userID int) ([]roleModels.Role, []roleModels.Permission, error) {
	return nil, nil, nil
}

type mockRBACService struct {
	rbac      *rbacModels.RBAC
	rbacError bool
}

func (m *mockRBACService) GetUserAccess(userID int) (*rbacModels.RBAC, error) {
	if m.rbacError {
		return nil, appErr.ErrInternal
	}
	if m.rbac == nil {
		return &rbacModels.RBAC{}, nil
	}
	return m.rbac, nil
}

func (m *mockRBACService) HasPermission(userID int, permissionName string) (bool, error) {
	return false, nil
}

func (m *mockRBACService) HasRole(userID int, roleName string) (bool, error) {
	return false, nil
}
