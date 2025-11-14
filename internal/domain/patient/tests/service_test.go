package patient_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/patient"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/patient/mocks"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/patient/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

func TestGetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := patient.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		expected := &models.Patient{
			ID:     1,
			Nombre: "Juan Perez",
			Sexo:   "M",
		}
		mockRepo.EXPECT().GetByID(1).Return(expected, nil)

		result, err := service.GetByID(1)
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		_, err := service.GetByID(0)
		require.Error(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(999).Return(nil, appErr.ErrNotFound)

		_, err := service.GetByID(999)
		require.Error(t, err)
	})
}

func TestGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := patient.NewService(mockRepo)

	expected := []models.Patient{
		{ID: 1, Nombre: "Juan Perez", Sexo: "M"},
		{ID: 2, Nombre: "Maria Lopez", Sexo: "F"},
	}

	mockRepo.EXPECT().GetAll().Return(expected, nil)

	result, err := service.GetAll()
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, expected, result)
}

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := patient.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		dto := &models.PatientCreateDTO{
			Nombre:          "Juan Perez",
			Sexo:            "M",
			FechaNacimiento: "1990-01-01",
		}

		mockRepo.EXPECT().Create(dto).Return(1, nil)

		id, err := service.Create(dto)
		require.NoError(t, err)
		require.Equal(t, 1, id)
	})

	t.Run("Nil DTO", func(t *testing.T) {
		_, err := service.Create(nil)
		require.Error(t, err)
	})

	t.Run("Empty Name", func(t *testing.T) {
		dto := &models.PatientCreateDTO{
			Nombre: "",
			Sexo:   "M",
		}

		_, err := service.Create(dto)
		require.Error(t, err)
	})

	t.Run("Empty Sexo", func(t *testing.T) {
		dto := &models.PatientCreateDTO{
			Nombre: "Juan Perez",
			Sexo:   "",
		}

		_, err := service.Create(dto)
		require.Error(t, err)
	})
}

func TestUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := patient.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		dto := &models.PatientUpdateDTO{
			Nombre:          "Juan Perez Updated",
			Sexo:            "M",
			FechaNacimiento: "1990-01-01",
		}

		mockRepo.EXPECT().Update(1, dto).Return(nil)

		err := service.Update(1, dto)
		require.NoError(t, err)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		dto := &models.PatientUpdateDTO{}

		err := service.Update(0, dto)
		require.Error(t, err)
	})

	t.Run("Nil DTO", func(t *testing.T) {
		err := service.Update(1, nil)
		require.Error(t, err)
	})
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := patient.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().Delete(1).Return(nil)

		err := service.Delete(1)
		require.NoError(t, err)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		err := service.Delete(0)
		require.Error(t, err)
	})
}

func TestSearchByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := patient.NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		expected := []models.PatientSearchResult{
			{ID: 1, Nombre: "Juan Perez", FechaNacimiento: "1990-01-01"},
		}

		mockRepo.EXPECT().SearchByName("Juan").Return(expected, nil)

		result, err := service.SearchByName("Juan")
		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, expected, result)
	})

	t.Run("Empty Name", func(t *testing.T) {
		_, err := service.SearchByName("")
		require.Error(t, err)
	})
}

// Helper functions
func strPtr(s string) *string {
	return &s
}
