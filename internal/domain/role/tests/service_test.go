package role_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/role"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/role/mocks"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/role/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// helper for creating service + mock
func setup(t *testing.T) (*mocks.MockRepository, role.Service, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockRepository(ctrl)
	svc := role.NewService(mockRepo)
	return mockRepo, svc, ctrl
}

func TestService_CreateRole(t *testing.T) {
	t.Parallel()

	t.Run("invalid input (nil role)", func(t *testing.T) {
		mockRepo, svc, ctrl := setup(t)
		defer ctrl.Finish()

		err := svc.CreateRole(nil)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))
		mockRepo.EXPECT().Create(gomock.Any()).Times(0)
	})

	t.Run("invalid input (empty name)", func(t *testing.T) {
		mockRepo, svc, ctrl := setup(t)
		defer ctrl.Finish()

		err := svc.CreateRole(&models.Role{Name: "", Description: "Handles patients"})
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))
		mockRepo.EXPECT().Create(gomock.Any()).Times(0)
	})

	t.Run("invalid input (empty description)", func(t *testing.T) {
		mockRepo, svc, ctrl := setup(t)
		defer ctrl.Finish()

		err := svc.CreateRole(&models.Role{Name: "Doctor", Description: ""})
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))
		mockRepo.EXPECT().Create(gomock.Any()).Times(0)
	})

	t.Run("successfully creates role", func(t *testing.T) {
		mockRepo, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().Create(gomock.Any()).Return(nil)

		err := svc.CreateRole(&models.Role{Name: "Doctor", Description: "Handles medical consultations"})
		require.NoError(t, err)
	})

	t.Run("repository returns already exists", func(t *testing.T) {
		mockRepo, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().
			Create(gomock.Any()).
			Return(appErr.Wrap("repo.Create", appErr.ErrAlreadyExists, errors.New("duplicate key")))

		err := svc.CreateRole(&models.Role{Name: "Admin", Description: "Full system access"})
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrAlreadyExists))
	})
}

func TestService_GetRoleByID(t *testing.T) {
	t.Parallel()

	t.Run("invalid ID", func(t *testing.T) {
		_, svc, ctrl := setup(t)
		defer ctrl.Finish()

		role, perms, err := svc.GetRoleByID(0)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))
		require.Nil(t, role)
		require.Nil(t, perms)
	})

	t.Run("repository returns not found", func(t *testing.T) {
		mockRepo, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().
			GetByID(99).
			Return(nil, appErr.Wrap("repo.GetByID", appErr.ErrNotFound, errors.New("no rows")))

		role, perms, err := svc.GetRoleByID(99)
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrNotFound))
		require.Nil(t, role)
		require.Nil(t, perms)
	})

	t.Run("successfully returns role and permissions", func(t *testing.T) {
		mockRepo, svc, ctrl := setup(t)
		defer ctrl.Finish()

		expectedRole := &models.Role{ID: 1, Name: "Admin", Description: "Full access"}
		expectedPerms := []models.Permission{{ID: 1, Name: "read-patient"}}

		mockRepo.EXPECT().GetByID(1).Return(expectedRole, nil)
		mockRepo.EXPECT().GetPermissions(1).Return(expectedPerms, nil)

		role, perms, err := svc.GetRoleByID(1)
		require.NoError(t, err)
		require.Equal(t, expectedRole, role)
		require.Equal(t, expectedPerms, perms)
	})
}

func TestService_UpdateRolePermissions(t *testing.T) {
	t.Parallel()

	t.Run("invalid ID", func(t *testing.T) {
		_, svc, ctrl := setup(t)
		defer ctrl.Finish()

		err := svc.UpdateRolePermissions(0, []int{1, 2})
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInvalidInput))
	})

	t.Run("clear permissions fails", func(t *testing.T) {
		mockRepo, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().
			ClearPermissions(1).
			Return(appErr.Wrap("repo.ClearPermissions", appErr.ErrInternal, errors.New("db error")))

		err := svc.UpdateRolePermissions(1, []int{1, 2})
		require.Error(t, err)
		require.True(t, errors.Is(err, appErr.ErrInternal))
	})

	t.Run("successfully updates permissions", func(t *testing.T) {
		mockRepo, svc, ctrl := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().ClearPermissions(1).Return(nil)
		mockRepo.EXPECT().AddPermission(1, 1).Return(nil)
		mockRepo.EXPECT().AddPermission(1, 2).Return(nil)

		err := svc.UpdateRolePermissions(1, []int{1, 2})
		require.NoError(t, err)
	})
}
