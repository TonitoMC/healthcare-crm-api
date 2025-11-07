package adapters

import (
	"time"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/schedule"
)

// ScheduleAdapter implements BusinessHoursValidator for appointment service
type ScheduleAdapter struct {
	Service schedule.Service
}

func NewScheduleAdapter(service schedule.Service) *ScheduleAdapter {
	return &ScheduleAdapter{Service: service}
}

func (s *ScheduleAdapter) IsWithinBusinessHours(date, start, end time.Time) (bool, error) {
	return s.Service.IsTimeRangeWithinWorkingHours(date, start, end)
}

func (s *ScheduleAdapter) GetEffectiveDay(date time.Time) (bool, error) {
	effectiveDay, err := s.Service.GetEffectiveDay(date)
	if err != nil {
		return false, err
	}
	return effectiveDay.Active, nil
}
