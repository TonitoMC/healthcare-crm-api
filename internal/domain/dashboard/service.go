package dashboard

import (
	"github.com/tonitomc/healthcare-crm-api/internal/domain/dashboard/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Service interface {
	GetStats() (*models.DashboardStats, error)
	GetRecentActivity(limit int) ([]models.RecentActivity, error)
	GetCriticalExams(limit int) ([]models.CriticalExam, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetStats() (*models.DashboardStats, error) {
	return s.repo.GetStats()
}

func (s *service) GetRecentActivity(limit int) ([]models.RecentActivity, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		return nil, appErr.Wrap("DashboardService.GetRecentActivity(limit too high)", appErr.ErrInvalidInput, nil)
	}
	return s.repo.GetRecentActivity(limit)
}

func (s *service) GetCriticalExams(limit int) ([]models.CriticalExam, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		return nil, appErr.Wrap("DashboardService.GetCriticalExams(limit too high)", appErr.ErrInvalidInput, nil)
	}
	return s.repo.GetCriticalExams(limit)
}
