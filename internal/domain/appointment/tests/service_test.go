package appointment_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/appointment"
	mockspackage "github.com/tonitomc/healthcare-crm-api/internal/domain/appointment/mocks"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/appointment/models"
	patientModels "github.com/tonitomc/healthcare-crm-api/internal/domain/patient/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

func TestGetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockspackage.NewMockRepository(ctrl)
	mockPatient := &mockPatientProvider{}
	mockSchedule := &mockScheduleValidator{}

	service := appointment.NewService(mockRepo, mockPatient, mockSchedule)

	t.Run("Success", func(t *testing.T) {
		expected := &models.Appointment{
			ID:         1,
			PacienteID: intPtr(1),
			Fecha:      time.Now(),
			Duracion:   1800,
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

func TestGetByDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockspackage.NewMockRepository(ctrl)
	mockPatient := &mockPatientProvider{}
	mockSchedule := &mockScheduleValidator{}

	service := appointment.NewService(mockRepo, mockPatient, mockSchedule)

	date := time.Date(2025, 11, 14, 0, 0, 0, 0, time.UTC)
	expected := []models.Appointment{
		{ID: 1, Fecha: date, Duracion: 1800},
		{ID: 2, Fecha: date.Add(2 * time.Hour), Duracion: 1800},
	}

	mockRepo.EXPECT().GetByDate(date).Return(expected, nil)

	result, err := service.GetByDate(date)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, expected, result)
}

func TestGetBetween(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockspackage.NewMockRepository(ctrl)
	mockPatient := &mockPatientProvider{}
	mockSchedule := &mockScheduleValidator{}

	service := appointment.NewService(mockRepo, mockPatient, mockSchedule)

	start := time.Date(2025, 11, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC)

	t.Run("Success", func(t *testing.T) {
		expected := []models.Appointment{
			{ID: 1, Fecha: start, Duracion: 1800},
		}
		mockRepo.EXPECT().GetBetween(start, end).Return(expected, nil)

		result, err := service.GetBetween(start, end)
		require.NoError(t, err)
		require.Len(t, result, 1)
	})

	t.Run("Invalid Range", func(t *testing.T) {
		_, err := service.GetBetween(end, start)
		require.Error(t, err)
	})
}

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockspackage.NewMockRepository(ctrl)
	mockPatient := &mockPatientProvider{exists: true}
	mockSchedule := &mockScheduleValidator{withinHours: true}

	service := appointment.NewService(mockRepo, mockPatient, mockSchedule)

	t.Run("Success with PacienteID", func(t *testing.T) {
		dto := &models.AppointmentCreateDTO{
			PacienteID: intPtr(1),
			Fecha:      time.Date(2025, 11, 14, 10, 0, 0, 0, time.UTC),
			Duracion:   1800,
		}

		mockRepo.EXPECT().GetBetween(gomock.Any(), gomock.Any()).Return([]models.Appointment{}, nil)
		mockRepo.EXPECT().Create(gomock.Any()).Return(1, nil)

		id, err := service.Create(dto)
		require.NoError(t, err)
		require.Equal(t, 1, id)
	})

	t.Run("Success with Nombre", func(t *testing.T) {
		dto := &models.AppointmentCreateDTO{
			Nombre:   strPtr("Walk-in Patient"),
			Fecha:    time.Date(2025, 11, 14, 10, 0, 0, 0, time.UTC),
			Duracion: 1800,
		}

		mockRepo.EXPECT().GetBetween(gomock.Any(), gomock.Any()).Return([]models.Appointment{}, nil)
		mockRepo.EXPECT().Create(gomock.Any()).Return(2, nil)

		id, err := service.Create(dto)
		require.NoError(t, err)
		require.Equal(t, 2, id)
	})

	t.Run("Missing Patient and Name", func(t *testing.T) {
		dto := &models.AppointmentCreateDTO{
			Fecha:    time.Date(2025, 11, 14, 10, 0, 0, 0, time.UTC),
			Duracion: 1800,
		}

		_, err := service.Create(dto)
		require.Error(t, err)
	})

	t.Run("Invalid Duration", func(t *testing.T) {
		dto := &models.AppointmentCreateDTO{
			PacienteID: intPtr(1),
			Fecha:      time.Date(2025, 11, 14, 10, 0, 0, 0, time.UTC),
			Duracion:   0,
		}

		_, err := service.Create(dto)
		require.Error(t, err)
	})

	t.Run("Patient Not Found", func(t *testing.T) {
		mockPatientNotFound := &mockPatientProvider{exists: false}
		svc := appointment.NewService(mockRepo, mockPatientNotFound, mockSchedule)

		dto := &models.AppointmentCreateDTO{
			PacienteID: intPtr(999),
			Fecha:      time.Date(2025, 11, 14, 10, 0, 0, 0, time.UTC),
			Duracion:   1800,
		}

		_, err := svc.Create(dto)
		require.Error(t, err)
	})

	t.Run("Outside Business Hours", func(t *testing.T) {
		mockScheduleOutside := &mockScheduleValidator{withinHours: false}
		svc := appointment.NewService(mockRepo, mockPatient, mockScheduleOutside)

		dto := &models.AppointmentCreateDTO{
			PacienteID: intPtr(1),
			Fecha:      time.Date(2025, 11, 14, 2, 0, 0, 0, time.UTC),
			Duracion:   1800,
		}

		_, err := svc.Create(dto)
		require.Error(t, err)
	})

	t.Run("Time Conflict", func(t *testing.T) {
		dto := &models.AppointmentCreateDTO{
			PacienteID: intPtr(1),
			Fecha:      time.Date(2025, 11, 14, 10, 0, 0, 0, time.UTC),
			Duracion:   1800,
		}

		existingAppt := []models.Appointment{
			{ID: 1, Fecha: time.Date(2025, 11, 14, 10, 15, 0, 0, time.UTC), Duracion: 1800},
		}

		mockRepo.EXPECT().GetBetween(gomock.Any(), gomock.Any()).Return(existingAppt, nil)

		_, err := service.Create(dto)
		require.Error(t, err)
	})
}

func TestUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockspackage.NewMockRepository(ctrl)
	mockPatient := &mockPatientProvider{}
	mockSchedule := &mockScheduleValidator{withinHours: true}

	service := appointment.NewService(mockRepo, mockPatient, mockSchedule)

	t.Run("Success", func(t *testing.T) {
		currentAppt := &models.Appointment{
			ID:       1,
			Fecha:    time.Date(2025, 11, 14, 10, 0, 0, 0, time.UTC),
			Duracion: 1800,
		}

		dto := &models.AppointmentUpdateDTO{
			Duracion: int64Ptr(2400),
		}

		mockRepo.EXPECT().GetByID(1).Return(currentAppt, nil)
		mockRepo.EXPECT().GetBetween(gomock.Any(), gomock.Any()).Return([]models.Appointment{}, nil)
		mockRepo.EXPECT().Update(1, dto).Return(nil)

		err := service.Update(1, dto)
		require.NoError(t, err)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		dto := &models.AppointmentUpdateDTO{}
		err := service.Update(0, dto)
		require.Error(t, err)
	})

	t.Run("Invalid Duration", func(t *testing.T) {
		dto := &models.AppointmentUpdateDTO{
			Duracion: int64Ptr(0),
		}
		err := service.Update(1, dto)
		require.Error(t, err)
	})
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockspackage.NewMockRepository(ctrl)
	mockPatient := &mockPatientProvider{}
	mockSchedule := &mockScheduleValidator{}

	service := appointment.NewService(mockRepo, mockPatient, mockSchedule)

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

func TestCreateWithNewPatient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockspackage.NewMockRepository(ctrl)
	mockPatient := &mockPatientProvider{exists: true}
	mockSchedule := &mockScheduleValidator{withinHours: true}

	service := appointment.NewService(mockRepo, mockPatient, mockSchedule)

	t.Run("Success", func(t *testing.T) {
		dto := &models.AppointmentWithNewPatientDTO{
			PatientData: patientModels.PatientCreateDTO{
				Nombre: "Juan Perez",
			},
		}
		dto.AppointmentData.Fecha = time.Date(2025, 11, 14, 10, 0, 0, 0, time.UTC)
		dto.AppointmentData.Duracion = 1800

		mockRepo.EXPECT().GetBetween(gomock.Any(), gomock.Any()).Return([]models.Appointment{}, nil)
		mockRepo.EXPECT().Create(gomock.Any()).Return(1, nil)

		id, err := service.CreateWithNewPatient(dto)
		require.NoError(t, err)
		require.Equal(t, 1, id)
	})

	t.Run("Invalid Duration", func(t *testing.T) {
		dto := &models.AppointmentWithNewPatientDTO{
			PatientData: patientModels.PatientCreateDTO{
				Nombre: "Juan Perez",
			},
		}
		dto.AppointmentData.Fecha = time.Date(2025, 11, 14, 10, 0, 0, 0, time.UTC)
		dto.AppointmentData.Duracion = 0

		_, err := service.CreateWithNewPatient(dto)
		require.Error(t, err)
	})

	t.Run("Patient Creation Failed", func(t *testing.T) {
		mockPatientFail := &mockPatientProvider{createError: true}
		svc := appointment.NewService(mockRepo, mockPatientFail, mockSchedule)

		dto := &models.AppointmentWithNewPatientDTO{
			PatientData: patientModels.PatientCreateDTO{
				Nombre: "Juan Perez",
			},
		}
		dto.AppointmentData.Fecha = time.Date(2025, 11, 14, 10, 0, 0, 0, time.UTC)
		dto.AppointmentData.Duracion = 1800

		_, err := svc.CreateWithNewPatient(dto)
		require.Error(t, err)
	})
}

func TestGetAvailableSlots(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockspackage.NewMockRepository(ctrl)
	mockPatient := &mockPatientProvider{}
	mockSchedule := &mockScheduleValidator{isOpen: true}

	service := appointment.NewService(mockRepo, mockPatient, mockSchedule)

	date := time.Date(2025, 11, 14, 0, 0, 0, 0, time.UTC)

	t.Run("Success with Mixed Availability", func(t *testing.T) {
		existingAppts := []models.Appointment{
			{ID: 1, Fecha: time.Date(2025, 11, 14, 9, 0, 0, 0, time.UTC), Duracion: 1800}, // 9:00-9:30
		}

		mockRepo.EXPECT().GetBetween(gomock.Any(), gomock.Any()).Return(existingAppts, nil)

		slots, err := service.GetAvailableSlots(date, 900) // 15-min slots
		require.NoError(t, err)
		require.NotEmpty(t, slots)

		// Check first slot (8:00-8:15) should be available
		require.True(t, slots[0].Available)
		require.Equal(t, 8, slots[0].Start.Hour())

		// Find a slot that overlaps with the 9:00 appointment
		foundUnavailable := false
		for _, slot := range slots {
			if slot.Start.Hour() == 9 && slot.Start.Minute() == 0 {
				require.False(t, slot.Available)
				foundUnavailable = true
				break
			}
		}
		require.True(t, foundUnavailable, "Should have found unavailable slot at 9:00")
	})

	t.Run("Closed Day", func(t *testing.T) {
		mockScheduleClosed := &mockScheduleValidator{isOpen: false}
		svc := appointment.NewService(mockRepo, mockPatient, mockScheduleClosed)

		slots, err := svc.GetAvailableSlots(date, 900)
		require.NoError(t, err)
		require.Empty(t, slots)
	})

	t.Run("Invalid Slot Duration Uses Default", func(t *testing.T) {
		mockRepo.EXPECT().GetBetween(gomock.Any(), gomock.Any()).Return([]models.Appointment{}, nil)

		slots, err := service.GetAvailableSlots(date, 0)
		require.NoError(t, err)
		require.NotEmpty(t, slots)
		// Default is 900 seconds (15 min)
		require.Equal(t, 900.0, slots[0].End.Sub(slots[0].Start).Seconds())
	})
}

func TestGetToday(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockspackage.NewMockRepository(ctrl)
	mockPatient := &mockPatientProvider{}
	mockSchedule := &mockScheduleValidator{}

	service := appointment.NewService(mockRepo, mockPatient, mockSchedule)

	expected := []models.Appointment{
		{ID: 1, Fecha: time.Now(), Duracion: 1800},
	}

	mockRepo.EXPECT().GetToday().Return(expected, nil)

	result, err := service.GetToday()
	require.NoError(t, err)
	require.Len(t, result, 1)
}

// Mock implementations
type mockPatientProvider struct {
	exists      bool
	createError bool
}

func (m *mockPatientProvider) GetByID(id int) (*patientModels.Patient, error) {
	return &patientModels.Patient{ID: id}, nil
}

func (m *mockPatientProvider) Exists(id int) (bool, error) {
	return m.exists, nil
}

func (m *mockPatientProvider) Create(dto *patientModels.PatientCreateDTO) (int, error) {
	if m.createError {
		return 0, appErr.ErrInternal
	}
	return 1, nil
}

type mockScheduleValidator struct {
	withinHours bool
	isOpen      bool
}

func (m *mockScheduleValidator) IsWithinBusinessHours(date, start, end time.Time) (bool, error) {
	return m.withinHours, nil
}

func (m *mockScheduleValidator) GetEffectiveDay(date time.Time) (bool, error) {
	return m.isOpen, nil
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}
