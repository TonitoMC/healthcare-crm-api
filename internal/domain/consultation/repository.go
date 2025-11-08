//go:generate mockgen -source=repository.go -destination=mocks/repository.go -package=mocks

package consultation

import (
	"database/sql"

	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/consultation/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Repository interface {
	// Consultations
	GetAll() ([]models.Consultation, error)
	GetByID(id int) (*models.Consultation, error)
	GetByPatient(patientID int) ([]models.Consultation, error)
	Create(consultation *models.Consultation) (int, error)
	Update(consultation *models.Consultation) error
	Delete(id int) error

	// --- Diagnostics ---
	GetDiagnosticsByConsultation(consultationID int) ([]models.Diagnostic, error)
	GetDiagnosticByID(id int) (*models.Diagnostic, error)
	CreateDiagnostic(d *models.Diagnostic) (int, error)
	UpdateDiagnostic(d *models.Diagnostic) error
	DeleteDiagnostic(id int) error

	// --- Treatments ---
	GetTreatmentsByDiagnostic(diagnosticID int) ([]models.Treatment, error)
	GetTreatmentByID(id int) (*models.Treatment, error)
	CreateTreatment(t *models.Treatment) (int, error)
	UpdateTreatment(t *models.Treatment) error
	DeleteTreatment(id int) error

	// --- Answers (Respuestas Cuestionarios) ---
	GetAnswersByConsultation(consultationID int) (*models.Answers, error)
	AddAnswers(a *models.Answers) (int, error)
	UpdateAnswers(a *models.Answers) error
	DeleteAnswers(consultationID int) error
}

type repository struct {
	db        *sql.DB
	validator QuestionnaireValidator
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAll() ([]models.Consultation, error) {
	rows, err := r.db.Query(`
		SELECT id, paciente_id, motivo, cuestionario_id, fecha, completada
		FROM consultas
		ORDER BY fecha DESC
	`)
	if err != nil {
		return nil, database.MapSQLError(err, "ConsultationRepository.GetAll")
	}
	defer rows.Close()

	var consultations []models.Consultation
	for rows.Next() {
		var c models.Consultation
		if err := rows.Scan(
			&c.ID,
			&c.PacienteID,
			&c.Motivo,
			&c.CuestionarioID,
			&c.Fecha,
			&c.Completada,
		); err != nil {
			return nil, appErr.Wrap("ConsultationRepository.GetAll(scan)", appErr.ErrInternal, err)
		}
		consultations = append(consultations, c)
	}

	return consultations, nil
}

func (r *repository) GetByID(id int) (*models.Consultation, error) {
	var c models.Consultation
	err := r.db.QueryRow(`
		SELECT id, paciente_id, motivo, cuestionario_id, fecha, completada
		FROM consultas
		WHERE id = $1
	`, id).Scan(&c.ID, &c.PacienteID, &c.Motivo, &c.CuestionarioID, &c.Fecha, &c.Completada)
	if err != nil {
		return nil, database.MapSQLError(err, "ConsultationRepository.GetByID")
	}

	return &c, nil
}

func (r *repository) GetByPatient(patientID int) ([]models.Consultation, error) {
	rows, err := r.db.Query(`
		SELECT id, paciente_id, motivo, cuestionario_id, fecha, completada
		FROM consultas
		WHERE paciente_id = $1
		ORDER BY fecha DESC
	`, patientID)
	if err != nil {
		return nil, database.MapSQLError(err, "ConsultationRepository.GetByPatient")
	}
	defer rows.Close()

	var consultations []models.Consultation
	for rows.Next() {
		var c models.Consultation
		if err := rows.Scan(&c.ID, &c.PacienteID, &c.Motivo, &c.CuestionarioID, &c.Fecha, &c.Completada); err != nil {
			return nil, appErr.Wrap("ConsultationRepository.GetByPatient(scan)", appErr.ErrInternal, err)
		}
		consultations = append(consultations, c)
	}

	return consultations, nil
}

func (r *repository) Create(consultation *models.Consultation) (int, error) {
	var id int
	err := r.db.QueryRow(`
		INSERT INTO consultas (paciente_id, motivo, cuestionario_id, fecha, completada)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, consultation.PacienteID, consultation.Motivo, consultation.CuestionarioID, consultation.Fecha, consultation.Completada).Scan(&id)
	if err != nil {
		return 0, database.MapSQLError(err, "ConsultationRepository.Create")
	}
	return id, nil
}

func (r *repository) Update(consultation *models.Consultation) error {
	res, err := r.db.Exec(`
		UPDATE consultas
		SET paciente_id = $1, motivo = $2, cuestionario_id = $3, fecha = $4, completada = $5
		WHERE id = $6
	`, consultation.PacienteID, consultation.Motivo, consultation.CuestionarioID, consultation.Fecha, consultation.Completada, consultation.ID)
	if err != nil {
		return database.MapSQLError(err, "ConsultationRepository.Update")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ConsultationRepository.Update", appErr.ErrNotFound, nil)
	}

	return nil
}

func (r *repository) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM consultas WHERE id = $1`, id)
	if err != nil {
		return database.MapSQLError(err, "ConsultationRepository.Delete")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ConsultationRepository.Delete", appErr.ErrNotFound, nil)
	}

	return nil
}

func (r *repository) GetDiagnosticsByConsultation(consultationID int) ([]models.Diagnostic, error) {
	rows, err := r.db.Query(`
		SELECT id, consulta_id, nombre, recomendacion
		FROM diagnosticos
		WHERE consulta_id = $1
	`, consultationID)
	if err != nil {
		return nil, database.MapSQLError(err, "ConsultationRepository.GetDiagnosticsByConsultation")
	}
	defer rows.Close()

	var diagnostics []models.Diagnostic
	for rows.Next() {
		var d models.Diagnostic
		if err := rows.Scan(
			&d.ID,
			&d.ConsultaID,
			&d.Nombre,
			&d.Recomendacion,
		); err != nil {
			return nil, appErr.Wrap("ConsultationRepository.GetDiagnosticsByConsultation(scan)", appErr.ErrInternal, err)
		}
		diagnostics = append(diagnostics, d)
	}
	return diagnostics, nil
}

func (r *repository) GetDiagnosticByID(id int) (*models.Diagnostic, error) {
	var d models.Diagnostic
	err := r.db.QueryRow(`
		SELECT id, consulta_id, nombre, recomendacion
		FROM diagnosticos
		WHERE id = $1
	`, id).Scan(
		&d.ID,
		&d.ConsultaID,
		&d.Nombre,
		&d.Recomendacion,
	)
	if err != nil {
		return nil, database.MapSQLError(err, "ConsultationRepository.GetDiagnosticByID")
	}
	return &d, nil
}

func (r *repository) CreateDiagnostic(d *models.Diagnostic) (int, error) {
	var id int
	err := r.db.QueryRow(`
		INSERT INTO diagnosticos (consulta_id, nombre, recomendacion)
		VALUES ($1, $2, $3)
		RETURNING id
	`, d.ConsultaID, d.Nombre, d.Recomendacion).Scan(&id)
	if err != nil {
		return 0, database.MapSQLError(err, "ConsultationRepository.CreateDiagnostic")
	}
	return id, nil
}

func (r *repository) UpdateDiagnostic(d *models.Diagnostic) error {
	res, err := r.db.Exec(`
		UPDATE diagnosticos
		SET nombre = $1, recomendacion = $2
		WHERE id = $3
	`, d.Nombre, d.Recomendacion, d.ID)
	if err != nil {
		return database.MapSQLError(err, "ConsultationRepository.UpdateDiagnostic")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ConsultationRepository.UpdateDiagnostic", appErr.ErrNotFound, nil)
	}

	return nil
}

func (r *repository) DeleteDiagnostic(id int) error {
	res, err := r.db.Exec(`DELETE FROM diagnosticos WHERE id = $1`, id)
	if err != nil {
		return database.MapSQLError(err, "ConsultationRepository.DeleteDiagnostic")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ConsultationRepository.DeleteDiagnostic", appErr.ErrNotFound, nil)
	}

	return nil
}

func (r *repository) GetTreatmentsByDiagnostic(diagnosticID int) ([]models.Treatment, error) {
	rows, err := r.db.Query(`
		SELECT id, nombre, diagnostico_id, componente_activo, presentacion, dosificacion, tiempo, frecuencia
		FROM tratamientos
		WHERE diagnostico_id = $1
	`, diagnosticID)
	if err != nil {
		return nil, database.MapSQLError(err, "ConsultationRepository.GetTreatmentsByDiagnostic")
	}
	defer rows.Close()

	var treatments []models.Treatment
	for rows.Next() {
		var t models.Treatment
		if err := rows.Scan(
			&t.ID,
			&t.Nombre,
			&t.DiagnosticoID,
			&t.ComponenteActivo,
			&t.Presentacion,
			&t.Dosificacion,
			&t.Tiempo,
			&t.Frecuencia,
		); err != nil {
			return nil, appErr.Wrap("ConsultationRepository.GetTreatmentsByDiagnostic(scan)", appErr.ErrInternal, err)
		}
		treatments = append(treatments, t)
	}

	return treatments, nil
}

func (r *repository) GetTreatmentByID(id int) (*models.Treatment, error) {
	var t models.Treatment
	err := r.db.QueryRow(`
		SELECT id, nombre, diagnostico_id, componente_activo, presentacion, dosificacion, tiempo, frecuencia
		FROM tratamientos
		WHERE id = $1
	`, id).Scan(
		&t.ID,
		&t.Nombre,
		&t.DiagnosticoID,
		&t.ComponenteActivo,
		&t.Presentacion,
		&t.Dosificacion,
		&t.Tiempo,
		&t.Frecuencia,
	)
	if err != nil {
		return nil, database.MapSQLError(err, "ConsultationRepository.GetTreatmentByID")
	}
	return &t, nil
}

func (r *repository) CreateTreatment(t *models.Treatment) (int, error) {
	var id int
	err := r.db.QueryRow(`
		INSERT INTO tratamientos (nombre, diagnostico_id, componente_activo, presentacion, dosificacion, tiempo, frecuencia)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, t.Nombre, t.DiagnosticoID, t.ComponenteActivo, t.Presentacion, t.Dosificacion, t.Tiempo, t.Frecuencia).Scan(&id)
	if err != nil {
		return 0, database.MapSQLError(err, "ConsultationRepository.CreateTreatment")
	}
	return id, nil
}

func (r *repository) UpdateTreatment(t *models.Treatment) error {
	res, err := r.db.Exec(`
		UPDATE tratamientos
		SET nombre = $1, componente_activo = $2, presentacion = $3, dosificacion = $4, tiempo = $5, frecuencia = $6
		WHERE id = $7
	`, t.Nombre, t.ComponenteActivo, t.Presentacion, t.Dosificacion, t.Tiempo, t.Frecuencia, t.ID)
	if err != nil {
		return database.MapSQLError(err, "ConsultationRepository.UpdateTreatment")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ConsultationRepository.UpdateTreatment", appErr.ErrNotFound, nil)
	}
	return nil
}

func (r *repository) DeleteTreatment(id int) error {
	res, err := r.db.Exec(`DELETE FROM tratamientos WHERE id = $1`, id)
	if err != nil {
		return database.MapSQLError(err, "ConsultationRepository.DeleteTreatment")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ConsultationRepository.DeleteTreatment", appErr.ErrNotFound, nil)
	}

	return nil
}

// --- ANSWERS IMPLEMENTATION ---

func (r *repository) GetAnswersByConsultation(consultationID int) (*models.Answers, error) {
	var a models.Answers
	err := r.db.QueryRow(`
		SELECT rc.id, rc.consulta_id, rc.cuestionario_id, rc.respuestas
		FROM respuestas_cuestionarios rc
		WHERE rc.consulta_id = $1
	`, consultationID).Scan(
		&a.ID,
		&a.ConsultaID,
		&a.CuestionarioID,
		&a.Respuestas,
	)
	if err != nil {
		return nil, database.MapSQLError(err, "ConsultationRepository.GetAnswersByConsultation")
	}
	return &a, nil
}

func (r *repository) AddAnswers(a *models.Answers) (int, error) {
	var id int
	err := r.db.QueryRow(`
		INSERT INTO respuestas_cuestionarios (consulta_id, cuestionario_id, respuestas)
		VALUES ($1, $2, $3)
		RETURNING id
	`, a.ConsultaID, a.CuestionarioID, a.Respuestas).Scan(&id)
	if err != nil {
		return 0, database.MapSQLError(err, "ConsultationRepository.AddAnswers")
	}
	return id, nil
}

func (r *repository) UpdateAnswers(a *models.Answers) error {
	res, err := r.db.Exec(`
		UPDATE respuestas_cuestionarios
		SET respuestas = $1
		WHERE consulta_id = $2
	`, a.Respuestas, a.ConsultaID)
	if err != nil {
		return database.MapSQLError(err, "ConsultationRepository.UpdateAnswers")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ConsultationRepository.UpdateAnswers", appErr.ErrNotFound, nil)
	}
	return nil
}

func (r *repository) DeleteAnswers(consultationID int) error {
	res, err := r.db.Exec(`
		DELETE FROM respuestas_cuestionarios
		WHERE consulta_id = $1
	`, consultationID)
	if err != nil {
		return database.MapSQLError(err, "ConsultationRepository.DeleteAnswers")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ConsultationRepository.DeleteAnswers", appErr.ErrNotFound, nil)
	}
	return nil
}
