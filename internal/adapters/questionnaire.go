package adapters

import (
	"encoding/json"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/questionnaire"
)

type QuestionnaireAdapter struct {
	Service questionnaire.Service
}

func NewQuestionnaireAdapter(service questionnaire.Service) *QuestionnaireAdapter {
	return &QuestionnaireAdapter{Service: service}
}

func (q *QuestionnaireAdapter) Validate(questionnaireID int, answers json.RawMessage) error {
	return q.Service.Validate(questionnaireID, answers)
}

