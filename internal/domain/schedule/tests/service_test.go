package tests

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
// Helper function to build simple time ranges
// -----------------------------------------------------------------------------
func mustParseTime(hour string) time.Time {
	t, _ := time.Parse("15:04", hour)
	return t
}

// -----------------------------------------------------------------------------
// Test Suite
// -----------------------------------------------------------------------------
func TestGetWorkingHours_SortsAndReturns(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	mockRepo.
		EXPECT().
		GetAllWorkingHours().
		Return([]models.WorkDay{
			{DayOfWeek: 3}, {DayOfWeek: 1}, {DayOfWeek: 2},
		}, nil)

	hours, err := svc.GetWorkingHours()
	assert.NoError(t, err)
	assert.Len(t, hours, 3)
	assert.Equal(t, 1, hours[0].DayOfWeek)
	assert.Equal(t, 3, hours[2].DayOfWeek)
}

func TestGetEffectiveDay_UsesSpecialOverride(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	date := time.Date(2025, 10, 30, 0, 0, 0, 0, time.UTC)

	special := &models.SpecialDay{
		Date: date, Active: true,
		Ranges: []models.TimeRange{{Start: mustParseTime("09:00"), End: mustParseTime("13:00")}},
	}

	mockRepo.
		EXPECT().
		GetSpecialHoursByDate(date).
		Return(special, nil)

	eff, err := svc.GetEffectiveDay(date)
	assert.NoError(t, err)
	assert.True(t, eff.IsOverride)
	assert.True(t, eff.Active)
	assert.Equal(t, 1, len(eff.Ranges))
}

func TestGetEffectiveDay_FallsBackToRegularIfNoSpecial(t *testing.T) {
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
			{DayOfWeek: 4, Active: true, Ranges: []models.TimeRange{{}}},
		}, nil)

	eff, err := svc.GetEffectiveDay(date)
	assert.NoError(t, err)
	assert.False(t, eff.IsOverride)
	assert.True(t, eff.Active)
}

func TestIsDateOpen_ClosedDay(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	date := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	special := &models.SpecialDay{Date: date, Active: false}

	mockRepo.
		EXPECT().
		GetSpecialHoursByDate(date).
		Return(special, nil)

	open, err := svc.IsDateOpen(date)
	assert.NoError(t, err)
	assert.False(t, open)
}

func TestIsTimeRangeWithinWorkingHours_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	date := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)
	workRange := models.TimeRange{Start: mustParseTime("09:00"), End: mustParseTime("13:00")}
	workday := models.WorkDay{DayOfWeek: 1, Active: true, Ranges: []models.TimeRange{workRange}}

	mockRepo.
		EXPECT().
		GetSpecialHoursByDate(date).
		Return(nil, nil)
	mockRepo.
		EXPECT().
		GetAllWorkingHours().
		Return([]models.WorkDay{workday}, nil)

	start := mustParseTime("09:30")
	end := mustParseTime("12:00")

	ok, err := svc.IsTimeRangeWithinWorkingHours(date, start, end)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestIsTimeRangeWithinWorkingHours_ClosedDay(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	date := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)
	special := &models.SpecialDay{Date: date, Active: false}

	mockRepo.
		EXPECT().
		GetSpecialHoursByDate(date).
		Return(special, nil)

	start := mustParseTime("09:00")
	end := mustParseTime("10:00")

	ok, err := svc.IsTimeRangeWithinWorkingHours(date, start, end)
	assert.False(t, ok)
	assert.Error(t, err)
	assert.True(t, appErr.IsDomainError(err))
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

func TestAddSpecialDay_ValidatesRanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := schedule.NewService(mockRepo)

	valid := models.SpecialDay{
		Date: time.Now(),
		Ranges: []models.TimeRange{
			{Start: mustParseTime("09:00"), End: mustParseTime("10:00")},
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
