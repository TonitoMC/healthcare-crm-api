//go:generate mockgen -source=service.go -destination=./mocks/service.go -package=mocks

package reminder

import (
	"time"

	models "github.com/tonitomc/healthcare-crm-api/internal/domain/reminder/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Service interface {
	Create(userID int, desc string, global bool) (*models.Reminder, error)
	GetForUser(userID int) ([]models.Reminder, error)
	SetDone(id int) error
	SetUndone(id int) error
	Delete(id int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(userID int, desc string, global bool) (*models.Reminder, error) {
	if desc == "" {
		return nil, appErr.Wrap("ReminderService.Create", appErr.ErrInvalidInput, nil)
	}

	// User-level reminder
	var uid *int
	if !global {
		uid = &userID
	}

	id, err := s.repo.Create(models.Reminder{
		UserID:      uid,
		Description: desc,
		Global:      global,
	})
	if err != nil {
		return nil, err
	}

	return &models.Reminder{
		ID:          id,
		UserID:      uid,
		Description: desc,
		Global:      global,
		CreatedAt:   time.Now(),
	}, nil
}

func (s *service) GetForUser(userID int) ([]models.Reminder, error) {
	return s.repo.GetForUser(userID)
}

func (s *service) SetDone(id int) error {
	return s.repo.MarkDone(id, time.Now())
}

func (s *service) SetUndone(id int) error {
	return s.repo.MarkUndone(id)
}

func (s *service) Delete(id int) error {
	return s.repo.Delete(id)
}
