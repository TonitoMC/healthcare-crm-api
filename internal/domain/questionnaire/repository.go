//go:generate mockgen -source=repository.go -destination=./mocks/repository.go -package=mocks

package questionnaire

import (
	"database/sql"

	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/questionnaire/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Repository interface {
	GetAll() ([]models.Questionnaire, error)
	GetByID(id int) (*models.Questionnaire, error)
	GetActiveByName(name string) (*models.Questionnaire, error)
	Create(q *models.Questionnaire) (int, error)
	Update(q *models.Questionnaire) error
	Delete(id int) error
	GetQuestionnaireNames() ([]string, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAll() ([]models.Questionnaire, error) {
	rows, err := r.db.Query(`
		SELECT id, nombre, version, activo, schema
		FROM cuestionarios
		ORDER BY nombre, version
	`)
	if err != nil {
		return nil, database.MapSQLError(err, "QuestionnaireRepository.GetAll")
	}
	defer rows.Close()

	var list []models.Questionnaire
	for rows.Next() {
		var q models.Questionnaire
		if err := rows.Scan(&q.ID, &q.Nombre, &q.Version, &q.Activo, &q.Schema); err != nil {
			return nil, appErr.Wrap("QuestionnaireRepository.GetAll(scan)", appErr.ErrInternal, err)
		}
		list = append(list, q)
	}

	return list, nil
}

func (r *repository) GetByID(id int) (*models.Questionnaire, error) {
	var q models.Questionnaire
	err := r.db.QueryRow(`
		SELECT id, nombre, version, activo, schema
		FROM cuestionarios
		WHERE id = $1
	`, id).Scan(&q.ID, &q.Nombre, &q.Version, &q.Activo, &q.Schema)
	if err != nil {
		return nil, database.MapSQLError(err, "QuestionnaireRepository.GetByID")
	}
	return &q, nil
}

func (r *repository) GetActiveByName(name string) (*models.Questionnaire, error) {
	var q models.Questionnaire
	err := r.db.QueryRow(`
		SELECT id, nombre, version, activo, schema
		FROM cuestionarios
		WHERE nombre = $1 AND activo = true
	`, name).Scan(&q.ID, &q.Nombre, &q.Version, &q.Activo, &q.Schema)
	if err != nil {
		return nil, database.MapSQLError(err, "QuestionnaireRepository.GetActiveByName")
	}
	return &q, nil
}

func (r *repository) Create(q *models.Questionnaire) (int, error) {
	var id int
	err := r.db.QueryRow(`
		INSERT INTO cuestionarios (nombre, version, activo, schema)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, q.Nombre, q.Version, q.Activo, q.Schema).Scan(&id)
	if err != nil {
		return 0, database.MapSQLError(err, "QuestionnaireRepository.Create")
	}
	return id, nil
}

func (r *repository) Update(q *models.Questionnaire) error {
	res, err := r.db.Exec(`
		UPDATE cuestionarios
		SET nombre = $1, version = $2, activo = $3, schema = $4
		WHERE id = $5
	`, q.Nombre, q.Version, q.Activo, q.Schema, q.ID)
	if err != nil {
		return database.MapSQLError(err, "QuestionnaireRepository.Update")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("QuestionnaireRepository.Update", appErr.ErrNotFound, nil)
	}
	return nil
}

func (r *repository) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM cuestionarios WHERE id = $1`, id)
	if err != nil {
		return database.MapSQLError(err, "QuestionnaireRepository.Delete")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("QuestionnaireRepository.Delete", appErr.ErrNotFound, nil)
	}
	return nil
}

func (r *repository) GetQuestionnaireNames() ([]string, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT nombre
		FROM cuestionarios
		ORDER BY nombre
	`)
	if err != nil {
		return nil, database.MapSQLError(err, "QuestionnaireRepository.GetQuestionnaireNames")
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, appErr.Wrap("QuestionnaireRepository.GetQuestionnaireNames(scan)", appErr.ErrInternal, err)
		}
		names = append(names, name)
	}

	return names, nil
}
