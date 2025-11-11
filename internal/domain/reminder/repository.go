//go:generate mockgen -source=repository.go -destination=./mocks/repository.go -package=mocks

package reminder

import (
	"database/sql"
	"time"

	dbErr "github.com/tonitomc/healthcare-crm-api/internal/database"
	models "github.com/tonitomc/healthcare-crm-api/internal/domain/reminder/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// Repository defines CRUD operations for reminders (recordatorios).
type Repository interface {
	// Reads
	GetAll() ([]models.Reminder, error)
	GetByID(id int) (*models.Reminder, error)
	GetByUser(userID int) ([]models.Reminder, error)
	GetPending() ([]models.Reminder, error)
	GetCompleted() ([]models.Reminder, error)

	// Writes
	Create(rem *models.ReminderCreateDTO) (int, error)
	MarkCompleted(id int, completedAt time.Time) error
	Delete(id int) error
}

// -----------------------------------------------------------------------------

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// -----------------------------------------------------------------------------
// Reads
// -----------------------------------------------------------------------------

func (r *repository) GetAll() ([]models.Reminder, error) {
	rows, err := r.db.Query(`
		SELECT id, usuario_id, descripcion, global, fecha_creacion, fecha_completado
		FROM recordatorios
		ORDER BY fecha_creacion DESC;
	`)
	if err != nil {
		return nil, dbErr.MapSQLError(err, "ReminderRepo.GetAll")
	}
	defer rows.Close()

	var result []models.Reminder
	for rows.Next() {
		var rem models.Reminder
		if err := rows.Scan(
			&rem.ID,
			&rem.UserID,
			&rem.Description,
			&rem.Global,
			&rem.CreatedAt,
			&rem.CompletedAt,
		); err != nil {
			return nil, appErr.Wrap("ReminderRepo.GetAll(scan)", appErr.ErrInternal, err)
		}
		result = append(result, rem)
	}

	return result, nil
}

func (r *repository) GetByID(id int) (*models.Reminder, error) {
	row := r.db.QueryRow(`
		SELECT id, usuario_id, descripcion, global, fecha_creacion, fecha_completado
		FROM recordatorios
		WHERE id = $1;
	`, id)

	var rem models.Reminder
	if err := row.Scan(
		&rem.ID,
		&rem.UserID,
		&rem.Description,
		&rem.Global,
		&rem.CreatedAt,
		&rem.CompletedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErr.NewDomainError(appErr.ErrNotFound, "Recordatorio no encontrado")
		}
		return nil, dbErr.MapSQLError(err, "ReminderRepo.GetByID")
	}
	return &rem, nil
}

func (r *repository) GetByUser(userID int) ([]models.Reminder, error) {
	rows, err := r.db.Query(`
		SELECT id, usuario_id, descripcion, global, fecha_creacion, fecha_completado
		FROM recordatorios
		WHERE usuario_id = $1 OR global = TRUE
		ORDER BY fecha_creacion DESC;
	`, userID)
	if err != nil {
		return nil, dbErr.MapSQLError(err, "ReminderRepo.GetByUser")
	}
	defer rows.Close()

	var result []models.Reminder
	for rows.Next() {
		var rem models.Reminder
		if err := rows.Scan(
			&rem.ID,
			&rem.UserID,
			&rem.Description,
			&rem.Global,
			&rem.CreatedAt,
			&rem.CompletedAt,
		); err != nil {
			return nil, appErr.Wrap("ReminderRepo.GetByUser(scan)", appErr.ErrInternal, err)
		}
		result = append(result, rem)
	}

	return result, nil
}

func (r *repository) GetPending() ([]models.Reminder, error) {
	rows, err := r.db.Query(`
		SELECT id, usuario_id, descripcion, global, fecha_creacion, fecha_completado
		FROM recordatorios
		WHERE fecha_completado IS NULL
		ORDER BY fecha_creacion DESC;
	`)
	if err != nil {
		return nil, dbErr.MapSQLError(err, "ReminderRepo.GetPending")
	}
	defer rows.Close()

	var result []models.Reminder
	for rows.Next() {
		var rem models.Reminder
		if err := rows.Scan(
			&rem.ID,
			&rem.UserID,
			&rem.Description,
			&rem.Global,
			&rem.CreatedAt,
			&rem.CompletedAt,
		); err != nil {
			return nil, appErr.Wrap("ReminderRepo.GetPending(scan)", appErr.ErrInternal, err)
		}
		result = append(result, rem)
	}

	return result, nil
}

func (r *repository) GetCompleted() ([]models.Reminder, error) {
	rows, err := r.db.Query(`
		SELECT id, usuario_id, descripcion, global, fecha_creacion, fecha_completado
		FROM recordatorios
		WHERE fecha_completado IS NOT NULL
		ORDER BY fecha_completado DESC;
	`)
	if err != nil {
		return nil, dbErr.MapSQLError(err, "ReminderRepo.GetCompleted")
	}
	defer rows.Close()

	var result []models.Reminder
	for rows.Next() {
		var rem models.Reminder
		if err := rows.Scan(
			&rem.ID,
			&rem.UserID,
			&rem.Description,
			&rem.Global,
			&rem.CreatedAt,
			&rem.CompletedAt,
		); err != nil {
			return nil, appErr.Wrap("ReminderRepo.GetCompleted(scan)", appErr.ErrInternal, err)
		}
		result = append(result, rem)
	}

	return result, nil
}

// -----------------------------------------------------------------------------
// Writes
// -----------------------------------------------------------------------------

func (r *repository) Create(rem *models.ReminderCreateDTO) (int, error) {
	if rem.Description == "" {
		return 0, appErr.Wrap("ReminderRepo.Create", appErr.ErrInvalidInput, nil)
	}

	var id int
	err := r.db.QueryRow(`
		INSERT INTO recordatorios (usuario_id, descripcion, global)
		VALUES ($1, $2, $3)
		RETURNING id;
	`, rem.UserID, rem.Description, rem.Global).Scan(&id)
	if err != nil {
		return 0, dbErr.MapSQLError(err, "ReminderRepo.Create")
	}

	return id, nil
}

func (r *repository) MarkCompleted(id int, completedAt time.Time) error {
	res, err := r.db.Exec(`
		UPDATE recordatorios
		SET fecha_completado = $1
		WHERE id = $2;
	`, completedAt, id)
	if err != nil {
		return dbErr.MapSQLError(err, "ReminderRepo.MarkCompleted")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.NewDomainError(appErr.ErrNotFound, "Recordatorio no encontrado")
	}
	return nil
}

func (r *repository) Delete(id int) error {
	res, err := r.db.Exec(`
		DELETE FROM recordatorios WHERE id = $1;
	`, id)
	if err != nil {
		return dbErr.MapSQLError(err, "ReminderRepo.Delete")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.NewDomainError(appErr.ErrNotFound, "Recordatorio no encontrado")
	}

	return nil
}
