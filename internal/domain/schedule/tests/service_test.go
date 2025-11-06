package schedule_test

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/schedule"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/schedule/mocks"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/schedule/models"
)

func makeTime(h, m int) time.Time {
	return time.Date(2025, 11, 2, h, m, 0, 0, time.UTC)
}

func TestGetWorkingHours_GroupsAndSortsCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)
	service := schedule.NewService(repo)

	repo.EXPECT().
		GetAllWorkingHours().
		Return([]models.WorkDay{
			{DayOfWeek: 1, Active: true, Ranges: []models.TimeRange{{Start: makeTime(10, 0), End: makeTime(11, 0)}}},
			{DayOfWeek: 1, Active: true, Ranges: []models.TimeRange{{Start: makeTime(8, 0), End: makeTime(9, 0)}}},
			{DayOfWeek: 2, Active: false},
		}, nil)

	result, err := service.GetWorkingHours()
	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, 1, result[0].DayOfWeek)
	assert.True(t, result[0].Active)
	assert.Equal(t, makeTime(8, 0), result[0].Ranges[0].Start)
	assert.Equal(t, makeTime(10, 0), result[0].Ranges[1].Start)
}

func TestGetWorkingHours_RepoErrorBubblesUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)
	service := schedule.NewService(repo)

	repo.EXPECT().
		GetAllWorkingHours().
		Return(nil, errors.New("db failure"))

	result, err := service.GetWorkingHours()
	assert.Nil(t, result)
	assert.EqualError(t, err, "db failure")
}

func TestAddSpecialDay_InvalidRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)
	service := schedule.NewService(repo)

	invalid := models.SpecialDay{
		Date:   time.Now(),
		Active: true,
		Ranges: []models.TimeRange{{Start: makeTime(10, 0), End: makeTime(9, 0)}}, // invalid
	}

	err := service.AddSpecialDay(invalid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Rango horario inválido")
}

func TestAddSpecialDay_ValidDelegatesToRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)
	service := schedule.NewService(repo)

	valid := models.SpecialDay{
		Date:   time.Now(),
		Active: true,
		Ranges: []models.TimeRange{{Start: makeTime(9, 0), End: makeTime(10, 0)}},
	}

	repo.EXPECT().UpdateSpecialHour(valid).Return(nil)
	err := service.AddSpecialDay(valid)
	assert.NoError(t, err)
}

func TestGetEffectiveDay_UsesSpecialOverride(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)
	service := schedule.NewService(repo)

	date := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)

	repo.EXPECT().
		GetSpecialHoursByDate(date).
		Return([]models.SpecialDay{
			{Date: date, Active: true, Ranges: []models.TimeRange{{Start: makeTime(8, 0), End: makeTime(9, 0)}}},
			{Date: date, Active: true, Ranges: []models.TimeRange{{Start: makeTime(10, 0), End: makeTime(11, 0)}}},
		}, nil)

	result, err := service.GetEffectiveDay(date)
	require.NoError(t, err)
	assert.True(t, result.IsOverride)
	assert.True(t, result.Active)
	assert.Len(t, result.Ranges, 2)
	assert.True(t, result.Ranges[0].Start.Before(result.Ranges[1].Start))
}

func TestGetEffectiveDay_FallbackToWorkingHours(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)
	service := schedule.NewService(repo)

	date := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC) // Monday (weekday=1)

	repo.EXPECT().GetSpecialHoursByDate(date).Return(nil, nil)
	repo.EXPECT().
		GetAllWorkingHours().
		Return([]models.WorkDay{
			{DayOfWeek: 1, Active: true, Ranges: []models.TimeRange{{Start: makeTime(9, 0), End: makeTime(17, 0)}}},
			{DayOfWeek: 2, Active: false},
		}, nil)

	result, err := service.GetEffectiveDay(date)
	require.NoError(t, err)
	assert.False(t, result.IsOverride)
	assert.True(t, result.Active)
	assert.Equal(t, 1, int(result.Date.Weekday()))
	assert.Len(t, result.Ranges, 1)
}

func TestIsTimeRangeWithinWorkingHours(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)
	service := schedule.NewService(repo)

	date := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)

	repo.EXPECT().
		GetSpecialHoursByDate(date).
		Return(nil, nil)
	repo.EXPECT().
		GetAllWorkingHours().
		Return([]models.WorkDay{
			{DayOfWeek: 1, Active: true, Ranges: []models.TimeRange{
				{Start: makeTime(9, 0), End: makeTime(17, 0)},
			}},
		}, nil)

	start := makeTime(10, 0)
	end := makeTime(11, 0)
	ok, err := service.IsTimeRangeWithinWorkingHours(date, start, end)
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestIsTimeRangeWithinWorkingHours_OutOfRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)
	service := schedule.NewService(repo)

	date := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)

	repo.EXPECT().GetSpecialHoursByDate(date).Return(nil, nil)
	repo.EXPECT().
		GetAllWorkingHours().
		Return([]models.WorkDay{
			{DayOfWeek: 1, Active: true, Ranges: []models.TimeRange{
				{Start: makeTime(9, 0), End: makeTime(17, 0)},
			}},
		}, nil)

	start := makeTime(18, 0)
	end := makeTime(19, 0)
	ok, err := service.IsTimeRangeWithinWorkingHours(date, start, end)
	assert.False(t, ok)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fuera del rango permitido")
}

func TestUpdateWorkDay_InvalidRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)
	service := schedule.NewService(repo)

	bad := models.WorkDay{
		DayOfWeek: 1,
		Active:    true,
		Ranges:    []models.TimeRange{{Start: makeTime(12, 0), End: makeTime(11, 0)}},
	}
	err := service.UpdateWorkDay(bad)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Rango horario inválido")
}

func TestUpdateWorkDay_ValidCallsRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)
	service := schedule.NewService(repo)

	good := models.WorkDay{
		DayOfWeek: 1,
		Active:    true,
		Ranges:    []models.TimeRange{{Start: makeTime(9, 0), End: makeTime(17, 0)}},
	}
	repo.EXPECT().UpdateWorkingHour(good).Return(nil)

	err := service.UpdateWorkDay(good)
	assert.NoError(t, err)
}

func TestDeleteSpecialDayByDate_CallsRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)
	service := schedule.NewService(repo)

	date := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	repo.EXPECT().DeleteSpecialHour(date).Return(nil)

	err := service.DeleteSpecialDay(date)
	assert.NoError(t, err)
}

func TestDeleteSpecialDayByDate_RepoErrorBubblesUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)
	service := schedule.NewService(repo)

	date := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	repo.EXPECT().DeleteSpecialHour(date).Return(errors.New("delete failed"))

	err := service.DeleteSpecialDay(date)
	assert.Error(t, err)
	assert.EqualError(t, err, "delete failed")
}
