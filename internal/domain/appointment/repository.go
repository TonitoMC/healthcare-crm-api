//go:generate mockgen -source=repository.go -destination=mocks/repository.go -package=mockspackage appointment

package appointment

import (
	"database/sql"
	"time"

	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/appointment/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Repository interface {
	GetByID(id int) (*models.Appointment, error)
	GetByDate(date time.Time) ([]models.Appointment, error)
	GetToday() ([]models.Appointment, error)
	GetBetween(start, end time.Time) ([]models.Appointment, error)
	Create(appt *models.AppointmentCreateDTO) (int, error)
	Update(id int, appt *models.AppointmentUpdateDTO) error
	Delete(id int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(id int) (*models.Appointment, error) {
	var a models.Appointment
	err := r.db.QueryRow(`
		SELECT c.id, c.paciente_id, c.nombre, c.fecha, c.duracion,
			   p.nombre, p.telefono, p.fecha_nacimiento
		FROM citas c
		LEFT JOIN pacientes p ON c.paciente_id = p.id
		WHERE c.id = $1
	`, id).Scan(
		&a.ID, &a.PacienteID, &a.Nombre, &a.Fecha, &a.Duracion,
		&a.NombrePaciente, &a.TelefonoPaciente, &a.FechaNacimiento,
	)
	if err != nil {
		return nil, database.MapSQLError(err, "AppointmentRepository.GetByID")
	}
	return &a, nil
}

func (r *repository) GetByDate(date time.Time) ([]models.Appointment, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	return r.GetBetween(startOfDay, endOfDay)
}

func (r *repository) GetToday() ([]models.Appointment, error) {
	return r.GetByDate(time.Now())
}

func (r *repository) GetBetween(start, end time.Time) ([]models.Appointment, error) {
	rows, err := r.db.Query(`
		SELECT c.id, c.paciente_id, c.nombre, c.fecha, c.duracion,
			   p.nombre, p.telefono, p.fecha_nacimiento
		FROM citas c
		LEFT JOIN pacientes p ON c.paciente_id = p.id
		WHERE c.fecha >= $1 AND c.fecha < $2
		ORDER BY c.fecha
	`, start, end)
	if err != nil {
		return nil, database.MapSQLError(err, "AppointmentRepository.GetBetween")
	}
	defer rows.Close()

	var appointments []models.Appointment
	for rows.Next() {
		var a models.Appointment
		if err := rows.Scan(
			&a.ID, &a.PacienteID, &a.Nombre, &a.Fecha, &a.Duracion,
			&a.NombrePaciente, &a.TelefonoPaciente, &a.FechaNacimiento,
		); err != nil {
			return nil, appErr.Wrap("AppointmentRepository.GetBetween(scan)", appErr.ErrInternal, err)
		}
		appointments = append(appointments, a)
	}
	return appointments, nil
}

func (r *repository) Create(appt *models.AppointmentCreateDTO) (int, error) {
	var id int
	err := r.db.QueryRow(`
		INSERT INTO citas (paciente_id, nombre, fecha, duracion)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, appt.PacienteID, appt.Nombre, appt.Fecha, appt.Duracion).Scan(&id)
	if err != nil {
		return 0, database.MapSQLError(err, "AppointmentRepository.Create")
	}
	return id, nil
}

func (r *repository) Update(id int, appt *models.AppointmentUpdateDTO) error {
	if appt.Fecha == nil && appt.Duracion == nil {
		return nil // nothing to update
	}

	query := "UPDATE citas SET "
	args := []interface{}{}
	argIdx := 1

	if appt.Fecha != nil {
		query += "fecha = $" + string(rune(argIdx+'0'))
		args = append(args, *appt.Fecha)
		argIdx++
	}
	if appt.Duracion != nil {
		if argIdx > 1 {
			query += ", "
		}
		query += "duracion = $" + string(rune(argIdx+'0'))
		args = append(args, *appt.Duracion)
		argIdx++
	}
	query += " WHERE id = $" + string(rune(argIdx+'0'))
	args = append(args, id)

	res, err := r.db.Exec(query, args...)
	if err != nil {
		return database.MapSQLError(err, "AppointmentRepository.Update")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("AppointmentRepository.Update", appErr.ErrNotFound, nil)
	}
	return nil
}

func (r *repository) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM citas WHERE id = $1`, id)
	if err != nil {
		return database.MapSQLError(err, "AppointmentRepository.Delete")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("AppointmentRepository.Delete", appErr.ErrNotFound, nil)
	}
	return nil
}
