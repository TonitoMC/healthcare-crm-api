//go:generate mockgen -source=repository.go -destination=./mocks/repository.go -package=mocks

package schedule

import (
	"database/sql"
	"time"

	dbErr "github.com/tonitomc/healthcare-crm-api/internal/database"
	models "github.com/tonitomc/healthcare-crm-api/internal/domain/schedule/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// Repository defines the data access contract for working hours and special days.
type Repository interface {
	GetAllWorkingHours() ([]models.WorkDay, error)
	GetAllSpecialHours() ([]models.SpecialDay, error)

	GetSpecialHoursBetween(start, end time.Time) ([]models.SpecialDay, error)
	GetSpecialHoursByDate(date time.Time) (*models.SpecialDay, error)

	CreateWorkingHour(day models.WorkDay) error
	UpdateWorkingHour(day models.WorkDay) error
	DeleteWorkingHour(id int) error

	CreateSpecialHour(day models.SpecialDay) error
	UpdateSpecialHour(day models.SpecialDay) error
	DeleteSpecialHour(id int) error
}

// -----------------------------------------------------------------------------
// Implementation
// -----------------------------------------------------------------------------

type repository struct {
	db *sql.DB
}

// NewRepository constructs a new schedule repository.
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// Working Hours
func (r *repository) GetAllWorkingHours() ([]models.WorkDay, error) {
	rows, err := r.db.Query(`
		SELECT id, dia_semana, hora_apertura, hora_cierre, abierto
		FROM horarios_laborales
		ORDER BY dia_semana, hora_apertura;
	`)
	if err != nil {
		return nil, dbErr.MapSQLError(err, "ScheduleRepo.GetAllWorkingHours")
	}
	defer rows.Close()

	var result []models.WorkDay
	for rows.Next() {
		var (
			id         int
			dayOfWeek  int
			start, end sql.NullTime
			active     bool
		)
		if err := rows.Scan(&id, &dayOfWeek, &start, &end, &active); err != nil {
			return nil, appErr.Wrap("ScheduleRepo.GetAllWorkingHours(scan)", appErr.ErrInternal, err)
		}

		wd := models.WorkDay{
			ID:        id,
			DayOfWeek: dayOfWeek,
			Active:    active,
		}

		if active && start.Valid && end.Valid {
			wd.Ranges = []models.TimeRange{{
				Start: start.Time,
				End:   end.Time,
			}}
		}

		result = append(result, wd)
	}

	return result, nil
}

func (r *repository) CreateWorkingHour(day models.WorkDay) error {
	if day.DayOfWeek < 1 || day.DayOfWeek > 7 {
		return appErr.Wrap("ScheduleRepo.CreateWorkingHour", appErr.ErrInvalidInput, nil)
	}

	var start, end *time.Time
	if day.Active && len(day.Ranges) > 0 {
		start = &day.Ranges[0].Start
		end = &day.Ranges[0].End
	}

	_, err := r.db.Exec(`
		INSERT INTO horarios_laborales (dia_semana, hora_apertura, hora_cierre, abierto)
		VALUES ($1, $2, $3, $4);
	`, day.DayOfWeek, start, end, day.Active)
	if err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.CreateWorkingHour")
	}
	return nil
}

func (r *repository) UpdateWorkingHour(day models.WorkDay) error {
	if day.ID <= 0 {
		return appErr.Wrap("ScheduleRepo.UpdateWorkingHour", appErr.ErrInvalidInput, nil)
	}

	var start, end *time.Time
	if day.Active && len(day.Ranges) > 0 {
		start = &day.Ranges[0].Start
		end = &day.Ranges[0].End
	}

	_, err := r.db.Exec(`
		UPDATE horarios_laborales
		SET dia_semana=$1, hora_apertura=$2, hora_cierre=$3, abierto=$4
		WHERE id=$5;
	`, day.DayOfWeek, start, end, day.Active, day.ID)
	if err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.UpdateWorkingHour")
	}
	return nil
}

func (r *repository) DeleteWorkingHour(id int) error {
	if id <= 0 {
		return appErr.Wrap("ScheduleRepo.DeleteWorkingHour", appErr.ErrInvalidInput, nil)
	}

	_, err := r.db.Exec(`DELETE FROM horarios_laborales WHERE id=$1;`, id)
	if err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.DeleteWorkingHour")
	}
	return nil
}

// Special Hours
func (r *repository) GetAllSpecialHours() ([]models.SpecialDay, error) {
	rows, err := r.db.Query(`
		SELECT id, fecha, hora_apertura, hora_cierre, abierto
		FROM horarios_especiales
		ORDER BY fecha;
	`)
	if err != nil {
		return nil, dbErr.MapSQLError(err, "ScheduleRepo.GetAllSpecialHours")
	}
	defer rows.Close()

	var result []models.SpecialDay
	for rows.Next() {
		var (
			id         int
			date       time.Time
			start, end sql.NullTime
			active     bool
		)
		if err := rows.Scan(&id, &date, &start, &end, &active); err != nil {
			return nil, appErr.Wrap("ScheduleRepo.GetAllSpecialHours(scan)", appErr.ErrInternal, err)
		}

		sd := models.SpecialDay{
			ID:     id,
			Date:   date,
			Active: active,
		}
		if active && start.Valid && end.Valid {
			sd.Ranges = []models.TimeRange{{
				Start: start.Time,
				End:   end.Time,
			}}
		}

		result = append(result, sd)
	}

	return result, nil
}

func (r *repository) GetSpecialHoursBetween(start, end time.Time) ([]models.SpecialDay, error) {
	rows, err := r.db.Query(`
		SELECT id, fecha, hora_apertura, hora_cierre, abierto
		FROM horarios_especiales
		WHERE fecha BETWEEN $1 AND $2
		ORDER BY fecha;
	`, start, end)
	if err != nil {
		return nil, dbErr.MapSQLError(err, "ScheduleRepo.GetSpecialHoursBetween")
	}
	defer rows.Close()

	var result []models.SpecialDay
	for rows.Next() {
		var (
			id          int
			date        time.Time
			open, close sql.NullTime
			active      bool
		)
		if err := rows.Scan(&id, &date, &open, &close, &active); err != nil {
			return nil, appErr.Wrap("ScheduleRepo.GetSpecialHoursBetween(scan)", appErr.ErrInternal, err)
		}

		day := models.SpecialDay{
			ID:     id,
			Date:   date,
			Active: active,
		}
		if active && open.Valid && close.Valid {
			day.Ranges = []models.TimeRange{{Start: open.Time, End: close.Time}}
		}

		result = append(result, day)
	}

	return result, nil
}

func (r *repository) GetSpecialHoursByDate(date time.Time) (*models.SpecialDay, error) {
	row := r.db.QueryRow(`
		SELECT id, fecha, hora_apertura, hora_cierre, abierto
		FROM horarios_especiales
		WHERE fecha = $1;
	`, date)

	var (
		id          int
		open, close sql.NullTime
		active      bool
	)
	var d time.Time
	if err := row.Scan(&id, &d, &open, &close, &active); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErr.Wrap("ScheduleRepo.GetSpecialHoursByDate", appErr.ErrNotFound, err)
		}
		return nil, dbErr.MapSQLError(err, "ScheduleRepo.GetSpecialHoursByDate")
	}

	day := &models.SpecialDay{
		ID:     id,
		Date:   d,
		Active: active,
	}
	if active && open.Valid && close.Valid {
		day.Ranges = []models.TimeRange{{Start: open.Time, End: close.Time}}
	}

	return day, nil
}

func (r *repository) CreateSpecialHour(day models.SpecialDay) error {
	_, err := r.db.Exec(`
		INSERT INTO horarios_especiales (fecha, hora_apertura, hora_cierre, abierto)
		VALUES ($1, $2, $3, $4);
	`, day.Date, getStart(day), getEnd(day), day.Active)
	if err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.CreateSpecialHour")
	}
	return nil
}

func (r *repository) UpdateSpecialHour(day models.SpecialDay) error {
	if day.ID <= 0 {
		return appErr.Wrap("ScheduleRepo.UpdateSpecialHour", appErr.ErrInvalidInput, nil)
	}

	_, err := r.db.Exec(`
		UPDATE horarios_especiales
		SET fecha=$1, hora_apertura=$2, hora_cierre=$3, abierto=$4
		WHERE id=$5;
	`, day.Date, getStart(day), getEnd(day), day.Active, day.ID)
	if err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.UpdateSpecialHour")
	}
	return nil
}

func (r *repository) DeleteSpecialHour(id int) error {
	if id <= 0 {
		return appErr.Wrap("ScheduleRepo.DeleteSpecialHour", appErr.ErrInvalidInput, nil)
	}

	_, err := r.db.Exec(`DELETE FROM horarios_especiales WHERE id=$1;`, id)
	if err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.DeleteSpecialHour")
	}
	return nil
}

// Helpers
func getStart(d any) *time.Time {
	switch v := d.(type) {
	case models.WorkDay:
		if v.Active && len(v.Ranges) > 0 {
			return &v.Ranges[0].Start
		}
	case models.SpecialDay:
		if v.Active && len(v.Ranges) > 0 {
			return &v.Ranges[0].Start
		}
	}
	return nil
}

func getEnd(d any) *time.Time {
	switch v := d.(type) {
	case models.WorkDay:
		if v.Active && len(v.Ranges) > 0 {
			return &v.Ranges[0].End
		}
	case models.SpecialDay:
		if v.Active && len(v.Ranges) > 0 {
			return &v.Ranges[0].End
		}
	}
	return nil
}
