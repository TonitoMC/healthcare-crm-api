package appointment

import (
	"time"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/appointment/models"
	patientModels "github.com/tonitomc/healthcare-crm-api/internal/domain/patient/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
	"github.com/tonitomc/healthcare-crm-api/pkg/timeutil"
)

// PatientProvider interface para evitar dependencias circulares
type PatientProvider interface {
	GetByID(id int) (*patientModels.Patient, error)
	Exists(id int) (bool, error)
	Create(dto *patientModels.PatientCreateDTO) (int, error)
}

// ScheduleValidator interface para validar horarios
type ScheduleValidator interface {
	IsWithinBusinessHours(date, start, end time.Time) (bool, error)
	GetEffectiveDay(date time.Time) (bool, error)
}

type Service interface {
	GetByID(id int) (*models.Appointment, error)
	GetByDate(date time.Time) ([]models.Appointment, error)
	GetToday() ([]models.Appointment, error)
	GetBetween(start, end time.Time) ([]models.Appointment, error)
	GetAvailableSlots(date time.Time, slotDuration int64) ([]models.AvailabilitySlot, error)
	Create(appt *models.AppointmentCreateDTO) (int, error)
	CreateWithNewPatient(dto *models.AppointmentWithNewPatientDTO) (int, error)
	Update(id int, appt *models.AppointmentUpdateDTO) error
	Delete(id int) error
}

type service struct {
	repo              Repository
	patientProvider   PatientProvider
	scheduleValidator ScheduleValidator
}

func NewService(repo Repository, patientProvider PatientProvider, scheduleValidator ScheduleValidator) Service {
	return &service{
		repo:              repo,
		patientProvider:   patientProvider,
		scheduleValidator: scheduleValidator,
	}
}

func (s *service) GetByID(id int) (*models.Appointment, error) {
	if id <= 0 {
		return nil, appErr.Wrap("AppointmentService.GetByID", appErr.ErrInvalidInput, nil)
	}
	return s.repo.GetByID(id)
}

func (s *service) GetByDate(date time.Time) ([]models.Appointment, error) {
	return s.repo.GetByDate(date)
}

func (s *service) GetToday() ([]models.Appointment, error) {
	return s.repo.GetToday()
}

func (s *service) GetBetween(start, end time.Time) ([]models.Appointment, error) {
	if start.After(end) {
		return nil, appErr.Wrap("AppointmentService.GetBetween(invalid range)", appErr.ErrInvalidInput, nil)
	}
	return s.repo.GetBetween(start, end)
}

func (s *service) Create(appt *models.AppointmentCreateDTO) (int, error) {
	if appt.PacienteID == nil && appt.Nombre == nil {
		return 0, appErr.Wrap("AppointmentService.Create(must provide paciente_id or nombre)", appErr.ErrInvalidInput, nil)
	}
	if appt.Duracion <= 0 {
		return 0, appErr.Wrap("AppointmentService.Create(duracion must be > 0)", appErr.ErrInvalidInput, nil)
	}

	appt.Fecha = timeutil.NormalizeToClinic(appt.Fecha)

	if appt.PacienteID != nil {
		exists, err := s.patientProvider.Exists(*appt.PacienteID)
		if err != nil {
			return 0, err
		}
		if !exists {
			return 0, appErr.Wrap("AppointmentService.Create(patient not found)", appErr.ErrNotFound, nil)
		}
	}

	endTime := appt.Fecha.Add(time.Duration(appt.Duracion) * time.Second)
	withinHours, err := s.scheduleValidator.IsWithinBusinessHours(appt.Fecha, appt.Fecha, endTime)
	if err != nil {
		return 0, err
	}
	if !withinHours {
		return 0, appErr.Wrap("AppointmentService.Create(time outside working hours)", appErr.ErrInvalidInput, nil)
	}

	const gapMinutes = 0
	dayStart := timeutil.StartOfClinicDay(appt.Fecha)
	dayEnd := dayStart.Add(24 * time.Hour)
	existing, err := s.repo.GetBetween(dayStart, dayEnd)
	if err != nil {
		return 0, err
	}

	endTimeWithGap := endTime.Add(time.Duration(gapMinutes) * time.Minute)
	for _, ex := range existing {
		exEnd := ex.Fecha.Add(time.Duration(ex.Duracion) * time.Second)
		exEndWithGap := exEnd.Add(time.Duration(gapMinutes) * time.Minute)
		if appt.Fecha.Before(exEndWithGap) && endTimeWithGap.After(ex.Fecha) {
			return 0, appErr.NewDomainError(appErr.ErrConflict, "El horario solicitado traslapa con otras citas")
		}
	}

	return s.repo.Create(appt)
}

func (s *service) CreateWithNewPatient(dto *models.AppointmentWithNewPatientDTO) (int, error) {
	if dto.AppointmentData.Duracion <= 0 {
		return 0, appErr.Wrap("AppointmentService.CreateWithNewPatient(duracion must be > 0)", appErr.ErrInvalidInput, nil)
	}

	patientID, err := s.patientProvider.Create(&dto.PatientData)
	if err != nil {
		return 0, err
	}

	appointmentDTO := &models.AppointmentCreateDTO{
		PacienteID: &patientID,
		Fecha:      dto.AppointmentData.Fecha,
		Duracion:   dto.AppointmentData.Duracion,
	}

	appointmentID, err := s.Create(appointmentDTO)
	if err != nil {
		return 0, err
	}

	return appointmentID, nil
}

func (s *service) GetAvailableSlots(date time.Time, slotDuration int64) ([]models.AvailabilitySlot, error) {
	if slotDuration <= 0 {
		slotDuration = 900 // 15 min default
	}

	isOpen, err := s.scheduleValidator.GetEffectiveDay(date)
	if err != nil {
		return nil, err
	}
	if !isOpen {
		return []models.AvailabilitySlot{}, nil
	}

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.Add(24 * time.Hour)
	appointments, err := s.repo.GetBetween(dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	var slots []models.AvailabilitySlot
	startTime := time.Date(date.Year(), date.Month(), date.Day(), 8, 0, 0, 0, date.Location())
	endTime := time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location())

	currentTime := startTime
	for currentTime.Before(endTime) {
		slotEnd := currentTime.Add(time.Duration(slotDuration) * time.Second)
		if slotEnd.After(endTime) {
			break
		}

		available := true
		for _, appt := range appointments {
			apptEnd := appt.Fecha.Add(time.Duration(appt.Duracion) * time.Second)
			if currentTime.Before(apptEnd) && slotEnd.After(appt.Fecha) {
				available = false
				break
			}
		}

		slots = append(slots, models.AvailabilitySlot{
			Start:     currentTime,
			End:       slotEnd,
			Available: available,
		})

		currentTime = slotEnd
	}

	return slots, nil
}

func (s *service) Update(id int, appt *models.AppointmentUpdateDTO) error {
	if id <= 0 {
		return appErr.Wrap("AppointmentService.Update(invalid id)", appErr.ErrInvalidInput, nil)
	}
	if appt.Duracion != nil && *appt.Duracion <= 0 {
		return appErr.Wrap("AppointmentService.Update(duracion must be > 0)", appErr.ErrInvalidInput, nil)
	}

	if appt.Fecha != nil || appt.Duracion != nil {
		current, err := s.repo.GetByID(id)
		if err != nil {
			return err
		}

		newFecha := current.Fecha
		if appt.Fecha != nil {
			newFecha = timeutil.NormalizeToClinic(*appt.Fecha)
		}
		newDuracion := current.Duracion
		if appt.Duracion != nil {
			newDuracion = *appt.Duracion
		}

		endTime := newFecha.Add(time.Duration(newDuracion) * time.Second)
		withinHours, err := s.scheduleValidator.IsWithinBusinessHours(newFecha, newFecha, endTime)
		if err != nil {
			return err
		}
		if !withinHours {
			return appErr.Wrap("AppointmentService.Update(time outside working hours)", appErr.ErrInvalidInput, nil)
		}

		const gapMinutes = 0
		dayStart := timeutil.StartOfClinicDay(newFecha)
		dayEnd := dayStart.Add(24 * time.Hour)
		existing, err := s.repo.GetBetween(dayStart, dayEnd)
		if err != nil {
			return err
		}

		endTimeWithGap := endTime.Add(time.Duration(gapMinutes) * time.Minute)
		for _, ex := range existing {
			if ex.ID == id {
				continue
			}
			exEnd := ex.Fecha.Add(time.Duration(ex.Duracion) * time.Second)
			exEndWithGap := exEnd.Add(time.Duration(gapMinutes) * time.Minute)
			if newFecha.Before(exEndWithGap) && endTimeWithGap.After(ex.Fecha) {
				return appErr.Wrap("AppointmentService.Update(time slot conflict)", appErr.ErrConflict, nil)
			}
		}

		if appt.Fecha != nil {
			*appt.Fecha = newFecha
		}
	}

	return s.repo.Update(id, appt)
}

func (s *service) Delete(id int) error {
	if id <= 0 {
		return appErr.Wrap("AppointmentService.Delete", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Delete(id)
}
