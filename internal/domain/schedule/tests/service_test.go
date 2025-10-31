package schedule_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/schedule"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/schedule/mocks"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/schedule/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------
func mustParseTime(hour string) time.Time {
	t, _ := time.Parse("15:04", hour)
	return t
}

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

func TestGetWorkingHours_GroupsAndSortsMultipleRanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	mockRepo.
		EXPECT().
		GetAllWorkingHours().
		Return([]models.WorkDay{
			{DayOfWeek: 1, Active: true, Ranges: []models.TimeRange{{Start: mustParseTime("15:00"), End: mustParseTime("18:00")}}},
			{DayOfWeek: 1, Active: true, Ranges: []models.TimeRange{{Start: mustParseTime("09:00"), End: mustParseTime("13:00")}}},
			{DayOfWeek: 2, Active: true, Ranges: []models.TimeRange{{Start: mustParseTime("08:00"), End: mustParseTime("12:00")}}},
		}, nil)

	hours, err := svc.GetWorkingHours()
	assert.NoError(t, err)
	assert.Len(t, hours, 2)
	assert.Equal(t, 1, hours[0].DayOfWeek)
	assert.Len(t, hours[0].Ranges, 2)
	assert.True(t, hours[0].Ranges[0].Start.Before(hours[0].Ranges[1].Start))
}

func TestGetSpecialHoursBetween_GroupsMultipleRangesPerDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	date := time.Date(2025, 11, 4, 0, 0, 0, 0, time.UTC)

	mockRepo.
		EXPECT().
		GetSpecialHoursBetween(gomock.Any(), gomock.Any()).
		Return([]models.SpecialDay{
			{Date: date, Active: true, Ranges: []models.TimeRange{{Start: mustParseTime("09:00"), End: mustParseTime("12:00")}}},
			{Date: date, Active: true, Ranges: []models.TimeRange{{Start: mustParseTime("14:00"), End: mustParseTime("18:00")}}},
		}, nil)

	specials, err := svc.GetSpecialHoursBetween(date, date)
	assert.NoError(t, err)
	assert.Len(t, specials, 1)
	assert.Len(t, specials[0].Ranges, 2)
	assert.True(t, specials[0].Ranges[0].Start.Before(specials[0].Ranges[1].Start))
}

func TestGetEffectiveDay_UsesMultipleSpecialRanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	date := time.Date(2025, 11, 10, 0, 0, 0, 0, time.UTC)

	mockRepo.
		EXPECT().
		GetSpecialHoursByDate(date).
		Return([]models.SpecialDay{
			{Date: date, Active: true, Ranges: []models.TimeRange{{Start: mustParseTime("09:00"), End: mustParseTime("12:00")}}},
			{Date: date, Active: true, Ranges: []models.TimeRange{{Start: mustParseTime("14:00"), End: mustParseTime("18:00")}}},
		}, nil)

	eff, err := svc.GetEffectiveDay(date)
	assert.NoError(t, err)
	assert.True(t, eff.IsOverride)
	assert.True(t, eff.Active)
	assert.Len(t, eff.Ranges, 2)
}

func TestGetEffectiveDay_MergesMultipleWorkingRanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	date := time.Date(2025, 10, 30, 0, 0, 0, 0, time.UTC) // Thursday = 4

	mockRepo.
		EXPECT().
		GetSpecialHoursByDate(date).
		Return(nil, nil)

	mockRepo.
		EXPECT().
		GetAllWorkingHours().
		Return([]models.WorkDay{
			{DayOfWeek: 4, Active: true, Ranges: []models.TimeRange{
				{Start: mustParseTime("09:00"), End: mustParseTime("12:00")},
				{Start: mustParseTime("14:00"), End: mustParseTime("17:00")},
			}},
		}, nil)

	eff, err := svc.GetEffectiveDay(date)
	assert.NoError(t, err)
	assert.False(t, eff.IsOverride)
	assert.True(t, eff.Active)
	assert.Len(t, eff.Ranges, 2)
}

func TestIsDateOpen_ClosedDay(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	date := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)

	mockRepo.
		EXPECT().
		GetSpecialHoursByDate(date).
		Return([]models.SpecialDay{
			{Date: date, Active: false},
		}, nil)

	open, err := svc.IsDateOpen(date)
	assert.NoError(t, err)
	assert.False(t, open)
}

func TestIsTimeRangeWithinWorkingHours_SuccessAcrossMultipleRanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	date := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)
	workday := models.WorkDay{
		DayOfWeek: 1,
		Active:    true,
		Ranges: []models.TimeRange{
			{Start: mustParseTime("09:00"), End: mustParseTime("13:00")},
			{Start: mustParseTime("15:00"), End: mustParseTime("18:00")},
		},
	}

	mockRepo.
		EXPECT().
		GetSpecialHoursByDate(date).
		Return(nil, nil)
	mockRepo.
		EXPECT().
		GetAllWorkingHours().
		Return([]models.WorkDay{workday}, nil)

	// Should pass (falls within second range)
	start := mustParseTime("15:30")
	end := mustParseTime("16:30")

	ok, err := svc.IsTimeRangeWithinWorkingHours(date, start, end)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUpdateWorkDay_InvalidRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	invalid := models.WorkDay{
		DayOfWeek: 1,
		Ranges: []models.TimeRange{
			{Start: mustParseTime("10:00"), End: mustParseTime("09:00")},
		},
	}

	err := svc.UpdateWorkDay(invalid)
	assert.Error(t, err)
	assert.True(t, appErr.IsDomainError(err))
}

func TestAddSpecialDay_ValidatesAndCreates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	valid := models.SpecialDay{
		Date: time.Now(),
		Ranges: []models.TimeRange{
			{Start: mustParseTime("09:00"), End: mustParseTime("10:00")},
			{Start: mustParseTime("11:00"), End: mustParseTime("12:00")},
		},
		Active: true,
	}

	mockRepo.
		EXPECT().
		CreateSpecialHour(valid).
		Return(nil)

	err := svc.AddSpecialDay(valid)
	assert.NoError(t, err)
}
