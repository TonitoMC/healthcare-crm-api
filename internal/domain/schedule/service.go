//go:generate mockgen -source=service.go -destination=mocks/service.go -package=mocks

package schedule

import (
	"sort"
	"time"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/schedule/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
	"github.com/tonitomc/healthcare-crm-api/pkg/timeutil"
)

// Service Interface
type Service interface {
	// Reads
	GetWorkingHours() ([]models.WorkDay, error)
	GetSpecialHoursBetween(start, end time.Time) ([]models.SpecialDay, error)
	GetEffectiveDay(date time.Time) (*models.EffectiveDay, error)
	GetEffectiveRange(start, end time.Time) ([]models.EffectiveDay, error)

	// Writes
	UpdateWorkDay(day models.WorkDay) error
	AddSpecialDay(day models.SpecialDay) error
	DeleteSpecialDay(date time.Time) error

	// Validations (Internal)
	IsTimeRangeWithinWorkingHours(date, start, end time.Time) (bool, error)
}

// Implementation
type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ============================================================================
// READ OPERATIONS
// ============================================================================

// GetWorkingHours returns all weekly recurring working days (Mon–Sun),
// grouping all time ranges belonging to the same weekday.
func (s *service) GetWorkingHours() ([]models.WorkDay, error) {
	raw, err := s.repo.GetAllWorkingHours()
	if err != nil {
		return nil, err
	}

	grouped := make(map[int]*models.WorkDay)
	for _, wd := range raw {
		if existing, ok := grouped[wd.DayOfWeek]; ok {
			existing.Ranges = append(existing.Ranges, wd.Ranges...)
			existing.Active = existing.Active || wd.Active // any active = active
		} else {
			grouped[wd.DayOfWeek] = &models.WorkDay{
				ID:        wd.ID,
				DayOfWeek: wd.DayOfWeek,
				Ranges:    wd.Ranges,
				Active:    wd.Active,
			}
		}
	}

	var merged []models.WorkDay
	for _, wd := range grouped {
		sort.Slice(wd.Ranges, func(i, j int) bool {
			return wd.Ranges[i].Start.Before(wd.Ranges[j].Start)
		})
		merged = append(merged, *wd)
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].DayOfWeek < merged[j].DayOfWeek
	})

	return merged, nil
}

// GetSpecialHoursBetween returns all special overrides in a date range,
// grouping all ranges for the same date.
func (s *service) GetSpecialHoursBetween(start, end time.Time) ([]models.SpecialDay, error) {
	raw, err := s.repo.GetSpecialHoursBetween(start, end)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string]*models.SpecialDay)
	for _, sd := range raw {
		key := sd.Date.Format("2006-01-02")
		if existing, ok := grouped[key]; ok {
			existing.Ranges = append(existing.Ranges, sd.Ranges...)
			existing.Active = existing.Active || sd.Active
		} else {
			grouped[key] = &models.SpecialDay{
				ID:     sd.ID,
				Date:   sd.Date,
				Ranges: sd.Ranges,
				Active: sd.Active,
			}
		}
	}

	var merged []models.SpecialDay
	for _, sd := range grouped {
		sort.Slice(sd.Ranges, func(i, j int) bool {
			return sd.Ranges[i].Start.Before(sd.Ranges[j].Start)
		})
		merged = append(merged, *sd)
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Date.Before(merged[j].Date)
	})

	return merged, nil
}

// GetEffectiveDay merges recurring + special schedules for a specific date.

func (s *service) GetEffectiveDay(date time.Time) (*models.EffectiveDay, error) {
	// --- 1. Check for special day overrides ---
	specials, err := s.repo.GetSpecialHoursByDate(date)
	if err != nil {
		return nil, err
	}

	if len(specials) > 0 {
		// Merge all special day entries for the same date (can have multiple ranges)
		var mergedRanges []models.TimeRange
		active := false

		for _, sd := range specials {
			if sd.Active {
				mergedRanges = append(mergedRanges, sd.Ranges...)
				active = true
			}
		}

		sort.Slice(mergedRanges, func(i, j int) bool {
			return mergedRanges[i].Start.Before(mergedRanges[j].Start)
		})

		return &models.EffectiveDay{
			Date:       specials[0].Date,
			Ranges:     mergedRanges,
			IsOverride: true,
			Active:     active,
		}, nil
	}

	// --- 2. Fallback: use recurring working hours if no special override exists ---
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	raw, err := s.repo.GetAllWorkingHours()
	if err != nil {
		return nil, err
	}

	var mergedRanges []models.TimeRange
	active := false

	for _, wd := range raw {
		if wd.DayOfWeek == weekday && wd.Active {
			mergedRanges = append(mergedRanges, wd.Ranges...)
			active = true
		}
	}

	sort.Slice(mergedRanges, func(i, j int) bool {
		return mergedRanges[i].Start.Before(mergedRanges[j].Start)
	})

	return &models.EffectiveDay{
		Date:       date,
		Ranges:     mergedRanges,
		IsOverride: false,
		Active:     active,
	}, nil
}

// GetEffectiveRange returns merged schedules for each date in a period,
// calling GetEffectiveDay for each date and aggregating results.
func (s *service) GetEffectiveRange(start, end time.Time) ([]models.EffectiveDay, error) {
	var days []models.EffectiveDay
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		eff, err := s.GetEffectiveDay(d)
		if err != nil {
			return nil, err
		}
		days = append(days, *eff)
	}
	return days, nil
}

// ============================================================================
// VALIDATION OPERATIONS
// ============================================================================

// IsTimeRangeWithinWorkingHours ensures an appointment fits within open slots.
func (s *service) IsTimeRangeWithinWorkingHours(date, start, end time.Time) (bool, error) {
	eff, err := s.GetEffectiveDay(date)
	if err != nil {
		return false, err
	}
	if !eff.Active {
		return false, appErr.NewDomainError(appErr.ErrConflict, "El día está cerrado.")
	}

	// Extract time-of-day from the appointment times (in local timezone)
	startTimeOfDay := timeutil.TimeOfDayMinutes(start.In(timeutil.ClinicLocation()))
	endTimeOfDay := timeutil.TimeOfDayMinutes(end.In(timeutil.ClinicLocation()))

	// Check if the appointment falls within any working range
	for _, r := range eff.Ranges {
		// Las franjas ya están ancladas con timezone de clínica
		rangeStartMinutes := timeutil.TimeOfDayMinutes(r.Start)
		rangeEndMinutes := timeutil.TimeOfDayMinutes(r.End)

		// Check if appointment fits within this range
		if startTimeOfDay >= rangeStartMinutes && endTimeOfDay <= rangeEndMinutes {
			return true, nil
		}
	}

	return false, appErr.NewDomainError(appErr.ErrConflict, "El horario solicitado está fuera del horario laboral.")
}

// ============================================================================
// WRITE OPERATIONS
// ============================================================================

func (s *service) UpdateWorkDay(day models.WorkDay) error {
	for _, r := range day.Ranges {
		if !r.IsValid() {
			return appErr.NewDomainError(appErr.ErrInvalidInput, "Rango horario inválido: hora de apertura mayor o igual a hora de cierre.")
		}
	}
	return s.repo.UpdateWorkingHour(day)
}

func (s *service) AddSpecialDay(day models.SpecialDay) error {
	for _, r := range day.Ranges {
		if !r.IsValid() {
			return appErr.NewDomainError(appErr.ErrInvalidInput, "Rango horario inválido: hora de apertura mayor o igual a hora de cierre.")
		}
	}
	return s.repo.UpdateSpecialHour(day)
}

func (s *service) UpdateSpecialDay(day models.SpecialDay) error {
	for _, r := range day.Ranges {
		if !r.IsValid() {
			return appErr.NewDomainError(appErr.ErrInvalidInput, "Rango horario inválido: hora de apertura mayor o igual a hora de cierre.")
		}
	}
	return s.repo.UpdateSpecialHour(day)
}

func (s *service) DeleteSpecialDay(date time.Time) error {
	return s.repo.DeleteSpecialHour(date)
}
