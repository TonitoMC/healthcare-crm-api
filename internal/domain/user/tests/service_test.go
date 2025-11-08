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

	t.Run("invalid input", func(t *testing.T) {
		_, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		require.ErrorIs(t, svc.CreateUser("", "mail@mail.com", "hash"), appErr.ErrInvalidInput)
		require.ErrorIs(t, svc.CreateUser("user", "", "hash"), appErr.ErrInvalidInput)
		require.ErrorIs(t, svc.CreateUser("user", "mail@mail.com", ""), appErr.ErrInvalidInput)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().Create(gomock.Any()).
			Return(appErr.Wrap("repo.Create", appErr.ErrInternal, errors.New("db fail")))

		err := svc.CreateUser("name", "mail@mail.com", "hash")
		require.ErrorIs(t, err, appErr.ErrInternal)
	})

	t.Run("success", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().Create(gomock.Any()).Return(nil)
		require.NoError(t, svc.CreateUser("ok", "ok@ok.com", "hash"))
	})
}

// -----------------------------------------------------------------------------
// GetAllUsers
// -----------------------------------------------------------------------------

func TestService_GetAllUsers(t *testing.T) {
	t.Parallel()

	t.Run("repository returns ErrNotFound", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().
			GetAll().
			Return(nil, appErr.Wrap("repo.GetAll", appErr.ErrNotFound, errors.New("no users")))

		users, err := svc.GetAllUsers()
		require.ErrorIs(t, err, appErr.ErrNotFound)
		require.Nil(t, users)
	})

	t.Run("successfully returns list", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		expected := []userModels.User{
			{ID: 1, Username: "admin", Email: "admin@example.com"},
			{ID: 2, Username: "secretary", Email: "sec@example.com"},
		}
		mockRepo.EXPECT().GetAll().Return(expected, nil)

		users, err := svc.GetAllUsers()
		require.NoError(t, err)
		require.Equal(t, expected, users)
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
		require.ErrorIs(t, err, appErr.ErrInvalidInput)
		require.Nil(t, u)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		mockRepo.EXPECT().
			GetByID(9).
			Return(nil, appErr.Wrap("repo.GetByID", appErr.ErrNotFound, errors.New("no rows")))
		u, err := svc.GetByID(9)
		require.ErrorIs(t, err, appErr.ErrNotFound)
		require.Nil(t, u)
	})

	t.Run("success", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		expected := &userModels.User{ID: 1, Username: "admin"}
		mockRepo.EXPECT().GetByID(1).Return(expected, nil)
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

	t.Run("invalid input", func(t *testing.T) {
		_, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		u, err := svc.GetByUsernameOrEmail("")
		require.ErrorIs(t, err, appErr.ErrInvalidInput)
		require.Nil(t, u)
	})

	t.Run("repository success", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		expected := &userModels.User{ID: 2, Username: "secretary"}
		mockRepo.EXPECT().GetByUsernameOrEmail("secretary").Return(expected, nil)
		u, err := svc.GetByUsernameOrEmail("secretary")
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
		require.ErrorIs(t, svc.UpdateUser(nil), appErr.ErrInvalidInput)
		require.ErrorIs(t, svc.UpdateUser(&userModels.User{ID: 0}), appErr.ErrInvalidInput)
	})

	t.Run("repository bubbles up", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		mockRepo.EXPECT().Update(gomock.Any()).
			Return(appErr.Wrap("repo.Update", appErr.ErrInternal, errors.New("db fail")))
		err := svc.UpdateUser(&userModels.User{ID: 1, Username: "x", Email: "y"})
		require.ErrorIs(t, err, appErr.ErrInternal)
	})

	t.Run("success", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		mockRepo.EXPECT().Update(gomock.Any()).Return(nil)
		err := svc.UpdateUser(&userModels.User{ID: 1, Username: "ok", Email: "ok@ok"})
		require.NoError(t, err)
	})
}

// -----------------------------------------------------------------------------
// DeleteUser
// -----------------------------------------------------------------------------

func TestService_DeleteUser(t *testing.T) {
	t.Parallel()

	t.Run("invalid id", func(t *testing.T) {
		_, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		require.ErrorIs(t, svc.DeleteUser(0), appErr.ErrInvalidInput)
	})

	t.Run("repository bubbles up", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		mockRepo.EXPECT().Delete(1).
			Return(appErr.Wrap("repo.Delete", appErr.ErrNotFound, errors.New("no rows")))
		err := svc.DeleteUser(1)
		require.ErrorIs(t, err, appErr.ErrNotFound)
	})

	t.Run("success", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		mockRepo.EXPECT().Delete(2).Return(nil)
		require.NoError(t, svc.DeleteUser(2))
	})
}

// -----------------------------------------------------------------------------
// Role Management: GetUserRoles / AddRole / RemoveRole / ClearRoles
// -----------------------------------------------------------------------------

func TestService_GetUserRoles(t *testing.T) {
	t.Parallel()

	t.Run("invalid id", func(t *testing.T) {
		_, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		r, err := svc.GetUserRoles(0)
		require.ErrorIs(t, err, appErr.ErrInvalidInput)
		require.Nil(t, r)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		mockRepo.EXPECT().GetByID(10).
			Return(nil, appErr.Wrap("repo.GetByID", appErr.ErrNotFound, errors.New("no user")))
		r, err := svc.GetUserRoles(10)
		require.ErrorIs(t, err, appErr.ErrNotFound)
		require.Nil(t, r)
	})

	t.Run("success", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		mockRepo.EXPECT().GetByID(1).
			Return(&userModels.User{ID: 1}, nil)
		expected := []roleModels.Role{{ID: 1, Name: "Admin"}}
		mockRepo.EXPECT().GetUserRoles(1).Return(expected, nil)
		r, err := svc.GetUserRoles(1)
		require.NoError(t, err)
		require.Equal(t, expected, r)
	})
}

func TestService_AddRole_RemoveRole_ClearRoles(t *testing.T) {
	t.Parallel()

	t.Run("AddRole validations and conflict", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()

		// invalid input
		require.ErrorIs(t, svc.AddRole(0, 1), appErr.ErrInvalidInput)

		// user not found
		mockRepo.EXPECT().GetByID(5).Return(nil, appErr.Wrap("repo.GetByID", appErr.ErrNotFound, errors.New("no user")))
		err := svc.AddRole(5, 1)
		require.ErrorIs(t, err, appErr.ErrNotFound)

		// duplicate role
		mockRepo.EXPECT().GetByID(1).Return(&userModels.User{ID: 1}, nil)
		mockRepo.EXPECT().GetUserRoles(1).Return([]roleModels.Role{{ID: 2}}, nil)
		err = svc.AddRole(1, 2)
		require.Error(t, err)
		require.Contains(t, err.Error(), "conflicto de datos")
		require.Contains(t, err.Error(), "ya tiene este rol")
	})

	t.Run("AddRole success flow", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		mockRepo.EXPECT().GetByID(1).Return(&userModels.User{ID: 1}, nil)
		mockRepo.EXPECT().GetUserRoles(1).Return([]roleModels.Role{}, nil)
		mockRepo.EXPECT().AddRole(1, 3).Return(nil)
		require.NoError(t, svc.AddRole(1, 3))
	})

	t.Run("RemoveRole & ClearRoles", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		mockRepo.EXPECT().GetByID(1).Return(&userModels.User{ID: 1}, nil)
		mockRepo.EXPECT().RemoveRole(1, 2).Return(nil)
		require.NoError(t, svc.RemoveRole(1, 2))

		mockRepo.EXPECT().GetByID(1).Return(&userModels.User{ID: 1}, nil)
		mockRepo.EXPECT().ClearRoles(1).Return(nil)
		require.NoError(t, svc.ClearRoles(1))
	})
}

// -----------------------------------------------------------------------------
// GetRolesAndPermissions
// -----------------------------------------------------------------------------

func TestService_GetRolesAndPermissions(t *testing.T) {
	t.Parallel()

	t.Run("invalid id", func(t *testing.T) {
		_, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		r, p, err := svc.GetRolesAndPermissions(0)
		require.ErrorIs(t, err, appErr.ErrInvalidInput)
		require.Nil(t, r)
		require.Nil(t, p)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo, _, svc, ctrl := setup(t)
		defer ctrl.Finish()
		mockRepo.EXPECT().GetByID(99).
			Return(nil, appErr.Wrap("repo.GetByID", appErr.ErrNotFound, errors.New("no user")))
		r, p, err := svc.GetRolesAndPermissions(99)
		require.ErrorIs(t, err, appErr.ErrNotFound)
		require.Nil(t, r)
		require.Nil(t, p)
	})

	t.Run("role service bubbles up", func(t *testing.T) {
		mockRepo, mockRoleSvc, svc, ctrl := setup(t)
		defer ctrl.Finish()
		mockRepo.EXPECT().GetByID(4).Return(&userModels.User{ID: 4}, nil)
		mockRepo.EXPECT().GetUserRoles(4).Return([]roleModels.Role{{ID: 1}}, nil)
		mockRoleSvc.EXPECT().GetPermissions(1).
			Return(nil, appErr.Wrap("role.GetPermissions", appErr.ErrInternal, errors.New("fail")))
		r, p, err := svc.GetRolesAndPermissions(4)
		require.ErrorIs(t, err, appErr.ErrInternal)
		require.Nil(t, r)
		require.Nil(t, p)
	})

	t.Run("deduplicates permissions", func(t *testing.T) {
		mockRepo, mockRoleSvc, svc, ctrl := setup(t)
		defer ctrl.Finish()
		mockRepo.EXPECT().GetByID(5).Return(&userModels.User{ID: 5}, nil)
		roles := []roleModels.Role{{ID: 1}, {ID: 2}}
		mockRepo.EXPECT().GetUserRoles(5).Return(roles, nil)
		mockRoleSvc.EXPECT().GetPermissions(1).Return([]roleModels.Permission{
			{ID: 1, Name: "read"},
		}, nil)
		mockRoleSvc.EXPECT().GetPermissions(2).Return([]roleModels.Permission{
			{ID: 1, Name: "read"},
			{ID: 2, Name: "write"},
		}, nil)
		_, perms, err := svc.GetRolesAndPermissions(5)
		require.NoError(t, err)
		require.Len(t, perms, 2)
	})
}
