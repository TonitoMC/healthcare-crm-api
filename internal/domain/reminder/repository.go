//go:generate mockgen -source=repository.go -destination=./mocks/repository.go -package=mocks

package reminder

import (
	"database/sql"
	"time"

	dbErr "github.com/tonitomc/healthcare-crm-api/internal/database"
	models "github.com/tonitomc/healthcare-crm-api/internal/domain/reminder/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Repository interface {
	Create(rem models.Reminder) (int, error)
	GetForUser(userID int) ([]models.Reminder, error)
	MarkDone(id int, completedAt time.Time) error
	MarkUndone(id int) error
	Delete(id int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// ----------------------------------------------------------------------

func (r *repository) Create(rem models.Reminder) (int, error) {
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

// ----------------------------------------------------------------------

func (r *repository) GetForUser(userID int) ([]models.Reminder, error) {
	rows, err := r.db.Query(`
		SELECT id, usuario_id, descripcion, global,
               fecha_creacion, fecha_completado
		FROM recordatorios
		WHERE global = TRUE OR usuario_id = $1
		ORDER BY fecha_creacion DESC;
	`, userID)
	if err != nil {
		return nil, dbErr.MapSQLError(err, "ReminderRepo.GetForUser")
	}
	defer rows.Close()

	var out []models.Reminder
	for rows.Next() {
		var rem models.Reminder
		var uid sql.NullInt32
		var completed sql.NullTime

		if err := rows.Scan(
			&rem.ID, &uid,
			&rem.Description, &rem.Global,
			&rem.CreatedAt, &completed,
		); err != nil {
			return nil, appErr.Wrap("ReminderRepo.Scan", appErr.ErrInternal, err)
		}

		if uid.Valid {
			val := int(uid.Int32)
			rem.UserID = &val
		}

		if completed.Valid {
			rem.CompletedAt = &completed.Time
		}

		out = append(out, rem)
	}
	return out, nil
}

// ----------------------------------------------------------------------

func (r *repository) MarkDone(id int, t time.Time) error {
	_, err := r.db.Exec(`
		UPDATE recordatorios
		SET fecha_completado = $1
		WHERE id = $2;
	`, t, id)
	return dbErr.MapSQLError(err, "ReminderRepo.MarkDone")
}

func (r *repository) MarkUndone(id int) error {
	_, err := r.db.Exec(`
		UPDATE recordatorios
		SET fecha_completado = NULL
		WHERE id = $1;
	`, id)
	return dbErr.MapSQLError(err, "ReminderRepo.MarkUndone")
}

func (r *repository) Delete(id int) error {
	_, err := r.db.Exec(`
		DELETE FROM recordatorios WHERE id = $1;
	`, id)
	return dbErr.MapSQLError(err, "ReminderRepo.Delete")
}
