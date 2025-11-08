//go:generate mockgen -source=service.go -destination=./mocks/service.go -package=mocks

package questionnaire

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/questionnaire/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Service interface {
	GetAll() ([]models.Questionnaire, error)
	GetByID(id int) (*models.Questionnaire, error)
	GetActiveByName(name string) (*models.Questionnaire, error)

	Create(dto *models.QuestionnaireCreateDTO) (int, error)
	Update(id int, dto *models.QuestionnaireUpdateDTO) error
	Delete(id int) error
	SetActive(id int) error
	SetInactive(id int) error
	GetQuestionnaireNames() ([]string, error)
	Validate(questionnaireID int, answers json.RawMessage) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAll() ([]models.Questionnaire, error) {
	return s.repo.GetAll()
}

func (s *service) GetByID(id int) (*models.Questionnaire, error) {
	if id <= 0 {
		return nil, appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del cuestionario es inválido.")
	}
	return s.repo.GetByID(id)
}

func (s *service) GetActiveByName(name string) (*models.Questionnaire, error) {
	if name == "" {
		return nil, appErr.NewDomainError(appErr.ErrInvalidInput, "El nombre del cuestionario es requerido.")
	}
	return s.repo.GetActiveByName(name)
}

func (s *service) Create(dto *models.QuestionnaireCreateDTO) (int, error) {
	if dto == nil {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "Datos inválidos para crear el cuestionario.")
	}

	if dto.Nombre == "" {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "El nombre del cuestionario es requerido.")
	}

	if dto.Version == "" {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "La versión del cuestionario es requerida.")
	}

	if err := validateSchemaStructure(dto.Schema); err != nil {
		return 0, err
	}

	// --- Enforce only one active version per name ---
	if dto.Activo {
		active, _ := s.repo.GetActiveByName(dto.Nombre)
		if active != nil {
			s.SetInactive(active.ID)
		}
	}

	q := &models.Questionnaire{
		Nombre:  dto.Nombre,
		Version: dto.Version,
		Activo:  dto.Activo,
		Schema:  dto.Schema,
	}

	id, err := s.repo.Create(q)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *service) Update(id int, dto *models.QuestionnaireUpdateDTO) error {
	if id <= 0 || dto == nil {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "Datos inválidos para actualizar el cuestionario.")
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if dto.Nombre == "" {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El nombre del cuestionario es requerido.")
	}
	if dto.Version == "" {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "La versión del cuestionario es requerida.")
	}

	if err := validateSchemaStructure(dto.Schema); err != nil {
		return err
	}

	// --- Rule: if setting activo=true, deactivate others ---
	if dto.Activo {
		active, _ := s.repo.GetActiveByName(dto.Nombre)
		if active != nil && active.ID != id {
			s.SetInactive(active.ID)
		}
	}

	existing.Nombre = dto.Nombre
	existing.Version = dto.Version
	existing.Activo = dto.Activo
	existing.Schema = dto.Schema

	if err := s.repo.Update(existing); err != nil {
		return err
	}

	return nil
}

func (s *service) Delete(id int) error {
	if id <= 0 {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del cuestionario es inválido.")
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// --- Optional business rule: prevent deleting active questionnaire ---
	if existing.Activo {
		return appErr.NewDomainError(appErr.ErrConflict, "No se puede eliminar un cuestionario activo. Desactívelo primero.")
	}

	return s.repo.Delete(id)
}

func validateSchemaStructure(schema json.RawMessage) error {
	var parsed struct {
		Questions []map[string]any `json:"questions"`
	}
	if err := json.Unmarshal(schema, &parsed); err != nil {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El esquema no es un JSON válido.")
	}

	if len(parsed.Questions) == 0 {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El esquema debe contener al menos una pregunta.")
	}

	validTypes := map[string]bool{"unilateral": true, "bilateral": true}
	validData := map[string]bool{"int": true, "float": true, "bool": true, "string": true}

	for i, q := range parsed.Questions {
		label, hasLabel := q["label"].(string)
		if !hasLabel || strings.TrimSpace(label) == "" {
			return appErr.NewDomainError(appErr.ErrInvalidInput,
				fmt.Sprintf("La pregunta %d no tiene un label válido.", i+1))
		}

		typ, ok1 := q["type"].(string)
		dt, ok2 := q["data_type"].(string)
		if !ok1 || !validTypes[typ] {
			return appErr.NewDomainError(appErr.ErrInvalidInput,
				fmt.Sprintf("La pregunta %d tiene un tipo inválido.", i+1))
		}
		if !ok2 || !validData[dt] {
			return appErr.NewDomainError(appErr.ErrInvalidInput,
				fmt.Sprintf("La pregunta %d tiene un tipo de dato inválido.", i+1))
		}

		if order, hasOrder := q["order"]; !hasOrder {
			return appErr.NewDomainError(appErr.ErrInvalidInput,
				fmt.Sprintf("La pregunta %d debe incluir un campo 'order'.", i+1))
		} else {
			switch order.(type) {
			case float64: // JSON numbers
			default:
				return appErr.NewDomainError(appErr.ErrInvalidInput,
					fmt.Sprintf("El campo 'order' de la pregunta %d debe ser numérico.", i+1))
			}
		}
	}

	return nil
}

func (s *service) GetQuestionnaireNames() ([]string, error) {
	names, err := s.repo.GetQuestionnaireNames()
	if err != nil {
		return nil, err
	}
	return names, nil
}

func (s *service) SetActive(id int) error {
	if id <= 0 {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El ID es inválido.")
	}

	q, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Deactivate any other active questionnaire with same name
	active, _ := s.repo.GetActiveByName(q.Nombre)
	if active != nil && active.ID != id {
		active.Activo = false
		if err := s.repo.Update(active); err != nil {
			return appErr.Wrap("QuestionnaireService.SetActive(deactivate)", appErr.ErrInternal, err)
		}
	}

	q.Activo = true
	return s.repo.Update(q)
}

func (s *service) SetInactive(id int) error {
	if id <= 0 {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El ID es inválido.")
	}

	q, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	q.Activo = false
	return s.repo.Update(q)
}

func (s *service) Validate(questionnaireID int, answers json.RawMessage) error {
	if questionnaireID <= 0 {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del cuestionario es inválido.")
	}
	if len(answers) == 0 {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "Las respuestas no pueden estar vacías.")
	}

	q, err := s.repo.GetByID(questionnaireID)
	if err != nil {
		return err
	}

	var schema struct {
		Questions []struct {
			Label    string `json:"label"`
			Type     string `json:"type"`
			DataType string `json:"data_type"`
			Order    int    `json:"order"`
		} `json:"questions"`
	}
	if err := json.Unmarshal(q.Schema, &schema); err != nil {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El esquema almacenado no es un JSON válido.")
	}

	var ans map[string]struct {
		Value   any    `json:"value"`
		Comment string `json:"comment"`
	}
	if err := json.Unmarshal(answers, &ans); err != nil {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "Las respuestas no son un JSON válido.")
	}

	for _, question := range schema.Questions {
		entry, exists := ans[question.Label]
		if !exists {
			return appErr.NewDomainError(appErr.ErrInvalidInput,
				fmt.Sprintf("Falta la respuesta para '%s'.", question.Label))
		}

		switch question.Type {
		case "bilateral":
			sides, ok := entry.Value.(map[string]any)
			if !ok {
				return appErr.NewDomainError(appErr.ErrInvalidInput,
					fmt.Sprintf("La respuesta para '%s' debe incluir los lados OI/OD.", question.Label))
			}
			for _, side := range []string{"OI", "OD"} {
				v, ok := sides[side]
				if !ok {
					return appErr.NewDomainError(appErr.ErrInvalidInput,
						fmt.Sprintf("Falta el valor de %s para '%s'.", side, question.Label))
				}
				if err := validateDataType(question.DataType, v); err != nil {
					return appErr.Wrap(fmt.Sprintf("Validación de '%s (%s)'", question.Label, side),
						appErr.ErrInvalidInput, err)
				}
			}

		case "unilateral":
			if err := validateDataType(question.DataType, entry.Value); err != nil {
				return appErr.Wrap(fmt.Sprintf("Validación de '%s'", question.Label),
					appErr.ErrInvalidInput, err)
			}

		default:
			return appErr.NewDomainError(appErr.ErrInvalidInput,
				fmt.Sprintf("Tipo '%s' inválido en el esquema para '%s'.", question.Type, question.Label))
		}

		// comment is optional but must exist as string
		if entry.Comment == "" {
			continue // can be empty string
		}
	}

	return nil
}

func validateDataType(expected string, val any) error {
	switch expected {
	case "int":
		if _, ok := val.(float64); !ok { // JSON numbers decode as float64
			return fmt.Errorf("se esperaba un número entero")
		}
	case "float":
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("se esperaba un número decimal")
		}
	case "bool":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("se esperaba un valor booleano")
		}
	case "string":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("se esperaba una cadena de texto")
		}
	default:
		return fmt.Errorf("tipo de dato no soportado: %s", expected)
	}
	return nil
}
