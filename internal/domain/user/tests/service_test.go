package tests

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	roleMocks "github.com/tonitomc/healthcare-crm-api/internal/domain/role/mocks"
	roleModels "github.com/tonitomc/healthcare-crm-api/internal/domain/role/models"
	user "github.com/tonitomc/healthcare-crm-api/internal/domain/user"
	userMocks "github.com/tonitomc/healthcare-crm-api/internal/domain/user/mocks"
	userModels "github.com/tonitomc/healthcare-crm-api/internal/domain/user/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func setup(t *testing.T) (*userMocks.MockRepository, *roleMocks.MockService, user.Service, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockRepo := userMocks.NewMockRepository(ctrl)
	mockRoleSvc := roleMocks.NewMockService(ctrl)
	svc := user.NewService(mockRepo, mockRoleSvc)
	return mockRepo, mockRoleSvc, svc, ctrl
}

// -----------------------------------------------------------------------------
// CreateUser
// -----------------------------------------------------------------------------

func TestService_CreateUser(t *testing.T) {
	t.Parallel()

	t.Run("invalid input (missing username/email/password)", func(t *testing.T) {
		mockRepo, mockRole, svc, ctrl := setup(t)
		defer ctrl.Finish()
		_ = mockRepo
		_ = mockRole

		err := svc.CreateUser("", "a@b.com", "hash")
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))

		err = svc.CreateUser("user", "", "hash")
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))

		err = svc.CreateUser("user", "a@b.com", "")
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))
	})

	t.Run("repository returns error", func(t *testing.T) {
		mockRepo, mockRole, svc, ctrl := setup(t)
		defer ctrl.Finish()
		_ = mockRole

		mockRepo.EXPECT().
			Create(gomock.Any()).
			Return(appErr.Wrap("repo.Create", appErr.ErrInternal, errors.New("db failure")))

		err := svc.CreateUser("doctor", "doc@example.com", "hash123")
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInternal))
	})

	t.Run("successfully creates user", func(t *testing.T) {
		mockRepo, mockRole, svc, ctrl := setup(t)
		defer ctrl.Finish()
		_ = mockRole

		mockRepo.EXPECT().
			Create(gomock.Any()).
			Return(nil)

		err := svc.CreateUser("nurse", "nurse@example.com", "securehash")
		require.NoError(t, err)
	})
}

// -----------------------------------------------------------------------------
// GetByID
// -----------------------------------------------------------------------------

func TestService_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("invalid ID", func(t *testing.T) {
		_, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		u, err := svc.GetByID(0)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))
		require.Nil(t, u)
	})

	t.Run("repository returns not found", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().
			GetByID(99).
			Return(nil, appErr.Wrap("repo.GetByID", appErr.ErrNotFound, errors.New("no rows")))

		u, err := svc.GetByID(99)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrNotFound))
		require.Nil(t, u)
	})

	t.Run("successfully retrieves user", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		expected := &userModels.User{ID: 1, Username: "admin", Email: "a@b.com"}
		mockRepo.EXPECT().
			GetByID(1).
			Return(expected, nil)

		u, err := svc.GetByID(1)
		require.NoError(t, err)
		require.Equal(t, expected, u)
	})
}

// -----------------------------------------------------------------------------
// GetByUsernameOrEmail
// -----------------------------------------------------------------------------

func TestService_GetByUsernameOrEmail(t *testing.T) {
	t.Parallel()

	t.Run("invalid identifier", func(t *testing.T) {
		_, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		u, err := svc.GetByUsernameOrEmail("")
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))
		require.Nil(t, u)
	})

	t.Run("repository returns user", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		expected := &userModels.User{ID: 1, Username: "doctor", Email: "doc@clinic.com"}
		mockRepo.EXPECT().
			GetByUsernameOrEmail("doctor").
			Return(expected, nil)

		u, err := svc.GetByUsernameOrEmail("doctor")
		require.NoError(t, err)
		require.Equal(t, expected, u)
	})
}

// -----------------------------------------------------------------------------
// UpdateUser
// -----------------------------------------------------------------------------

func TestService_UpdateUser(t *testing.T) {
	t.Parallel()

	t.Run("invalid input", func(t *testing.T) {
		_, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		err := svc.UpdateUser(nil)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))

		err = svc.UpdateUser(&userModels.User{ID: 0})
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().
			Update(gomock.Any()).
			Return(appErr.Wrap("repo.Update", appErr.ErrInternal, errors.New("db failure")))

		err := svc.UpdateUser(&userModels.User{ID: 1, Username: "x", Email: "x@x.com"})
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInternal))
	})

	t.Run("successfully updates user", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().Update(gomock.Any()).Return(nil)

		err := svc.UpdateUser(&userModels.User{ID: 1, Username: "updated", Email: "new@mail.com"})
		require.NoError(t, err)
	})
}

// -----------------------------------------------------------------------------
// DeleteUser
// -----------------------------------------------------------------------------

func TestService_DeleteUser(t *testing.T) {
	t.Parallel()

	t.Run("invalid ID", func(t *testing.T) {
		_, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		err := svc.DeleteUser(0)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))
	})

	t.Run("repository not found", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().
			Delete(99).
			Return(appErr.Wrap("repo.Delete", appErr.ErrNotFound, errors.New("no rows")))

		err := svc.DeleteUser(99)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrNotFound))
	})

	t.Run("successfully deletes user", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().
			Delete(1).
			Return(nil)

		err := svc.DeleteUser(1)
		require.NoError(t, err)
	})
}

// -----------------------------------------------------------------------------
// Role Management
// -----------------------------------------------------------------------------

func TestService_GetUserRoles(t *testing.T) {
	t.Parallel()

	t.Run("invalid user ID", func(t *testing.T) {
		_, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		roles, err := svc.GetUserRoles(0)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))
		require.Nil(t, roles)
	})

	t.Run("repository returns roles", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		expected := []roleModels.Role{
			{ID: 1, Name: "Admin", Description: "System administrator"},
		}
		mockRepo.EXPECT().
			GetUserRoles(1).
			Return(expected, nil)

		roles, err := svc.GetUserRoles(1)
		require.NoError(t, err)
		require.Equal(t, expected, roles)
	})
}

func TestService_AddRole_RemoveRole_ClearRoles(t *testing.T) {
	t.Parallel()

	t.Run("invalid inputs", func(t *testing.T) {
		_, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		require.Error(t, svc.AddRole(0, 1))
		require.Error(t, svc.RemoveRole(1, 0))
		require.Error(t, svc.ClearRoles(0))
	})

	t.Run("repository bubbles up", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().AddRole(1, 2).Return(appErr.Wrap("repo.AddRole", appErr.ErrInternal, errors.New("fail")))
		err := svc.AddRole(1, 2)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInternal))

		mockRepo.EXPECT().RemoveRole(1, 2).Return(appErr.Wrap("repo.RemoveRole", appErr.ErrNotFound, errors.New("no row")))
		err = svc.RemoveRole(1, 2)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrNotFound))

		mockRepo.EXPECT().ClearRoles(1).Return(appErr.Wrap("repo.ClearRoles", appErr.ErrInternal, errors.New("fail")))
		err = svc.ClearRoles(1)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInternal))
	})

	t.Run("success flows", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().AddRole(1, 2).Return(nil)
		require.NoError(t, svc.AddRole(1, 2))

		mockRepo.EXPECT().RemoveRole(1, 2).Return(nil)
		require.NoError(t, svc.RemoveRole(1, 2))

		mockRepo.EXPECT().ClearRoles(1).Return(nil)
		require.NoError(t, svc.ClearRoles(1))
	})
}

// -----------------------------------------------------------------------------
// GetRolesAndPermissions
// -----------------------------------------------------------------------------

func TestService_GetRolesAndPermissions(t *testing.T) {
	t.Parallel()

	t.Run("invalid user ID", func(t *testing.T) {
		_, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		r, p, err := svc.GetRolesAndPermissions(0)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))
		require.Nil(t, r)
		require.Nil(t, p)
	})

	t.Run("repository bubbles up", func(t *testing.T) {
		mockRepo, mockRoleSvc, svc, ctrl := setup(t)
		defer ctrl.Finish()
		_ = mockRoleSvc

		mockRepo.EXPECT().
			GetUserRoles(10).
			Return(nil, appErr.Wrap("repo.GetUserRoles", appErr.ErrInternal, errors.New("fail")))

		r, p, err := svc.GetRolesAndPermissions(10)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInternal))
		require.Nil(t, r)
		require.Nil(t, p)
	})

	t.Run("role service returns permissions correctly", func(t *testing.T) {
		mockRepo, mockRoleSvc, svc, ctrl := setup(t)
		defer ctrl.Finish()

		roles := []roleModels.Role{
			{ID: 1, Name: "Doctor", Description: "Medical staff"},
			{ID: 2, Name: "Admin", Description: "System administrator"},
		}
		mockRepo.EXPECT().GetUserRoles(5).Return(roles, nil)

		mockRoleSvc.EXPECT().
			GetPermissions(1).
			Return([]roleModels.Permission{
				{ID: 1, Name: "read", Description: "Can read patient records"},
			}, nil)
		mockRoleSvc.EXPECT().
			GetPermissions(2).
			Return([]roleModels.Permission{
				{ID: 2, Name: "write", Description: "Can edit patient records"},
			}, nil)

		r, p, err := svc.GetRolesAndPermissions(5)
		require.NoError(t, err)
		require.Len(t, r, 2)
		require.Len(t, p, 2)
	})

	t.Run("duplicate permissions deduplicated", func(t *testing.T) {
		mockRepo, mockRoleSvc, svc, ctrl := setup(t)
		defer ctrl.Finish()

		roles := []roleModels.Role{
			{ID: 1, Name: "Admin", Description: "System administrator"},
			{ID: 2, Name: "SuperAdmin", Description: "Elevated access"},
		}
		mockRepo.EXPECT().GetUserRoles(1).Return(roles, nil)

		mockRoleSvc.EXPECT().
			GetPermissions(1).
			Return([]roleModels.Permission{
				{ID: 1, Name: "read", Description: "Can read data"},
			}, nil)
		mockRoleSvc.EXPECT().
			GetPermissions(2).
			Return([]roleModels.Permission{
				{ID: 1, Name: "read", Description: "Can read data"},
				{ID: 2, Name: "write", Description: "Can modify data"},
			}, nil)

		_, perms, err := svc.GetRolesAndPermissions(1)
		require.NoError(t, err)
		require.Len(t, perms, 2) // should deduplicate "read"
	})
}
