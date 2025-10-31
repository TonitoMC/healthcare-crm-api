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
	GetSpecialHoursByDate(date time.Time) ([]models.SpecialDay, error)

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

// -----------------------------------------------------------------------------
// Working Hours
// -----------------------------------------------------------------------------

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
			id        int
			dayOfWeek int
			openStr   sql.NullString
			closeStr  sql.NullString
			active    bool
		)
		if err := rows.Scan(&id, &dayOfWeek, &openStr, &closeStr, &active); err != nil {
			return nil, appErr.Wrap("ScheduleRepo.GetAllWorkingHours(scan)", appErr.ErrInternal, err)
		}

		wd := models.WorkDay{
			ID:        id,
			DayOfWeek: dayOfWeek,
			Active:    active,
		}

		if active && openStr.Valid && closeStr.Valid {
			start, err1 := time.Parse("15:04:05", openStr.String)
			end, err2 := time.Parse("15:04:05", closeStr.String)
			if err1 == nil && err2 == nil {
				wd.Ranges = []models.TimeRange{{
					Start: start,
					End:   end,
				}}
			}
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

// -----------------------------------------------------------------------------
// Special Hours
// -----------------------------------------------------------------------------

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
			id       int
			date     time.Time
			openStr  sql.NullString
			closeStr sql.NullString
			active   bool
		)
		if err := rows.Scan(&id, &date, &openStr, &closeStr, &active); err != nil {
			return nil, appErr.Wrap("ScheduleRepo.GetAllSpecialHours(scan)", appErr.ErrInternal, err)
		}

		sd := models.SpecialDay{
			ID:     id,
			Date:   date,
			Active: active,
		}
		if active && openStr.Valid && closeStr.Valid {
			start, err1 := time.Parse("15:04:05", openStr.String)
			end, err2 := time.Parse("15:04:05", closeStr.String)
			if err1 == nil && err2 == nil {
				sd.Ranges = []models.TimeRange{{
					Start: start,
					End:   end,
				}}
			}
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
			id       int
			date     time.Time
			openStr  sql.NullString
			closeStr sql.NullString
			active   bool
		)
		if err := rows.Scan(&id, &date, &openStr, &closeStr, &active); err != nil {
			return nil, appErr.Wrap("ScheduleRepo.GetSpecialHoursBetween(scan)", appErr.ErrInternal, err)
		}

		sd := models.SpecialDay{
			ID:     id,
			Date:   date,
			Active: active,
		}
		if active && openStr.Valid && closeStr.Valid {
			start, err1 := time.Parse("15:04:05", openStr.String)
			end, err2 := time.Parse("15:04:05", closeStr.String)
			if err1 == nil && err2 == nil {
				sd.Ranges = []models.TimeRange{{Start: start, End: end}}
			}
		}

		result = append(result, sd)
	}

	return result, nil
}

// GetSpecialHoursByDate returns all special hour entries for a specific date.
// Multiple rows can exist (e.g., morning + afternoon shifts).
func (r *repository) GetSpecialHoursByDate(date time.Time) ([]models.SpecialDay, error) {
	rows, err := r.db.Query(`
		SELECT id, fecha, hora_apertura, hora_cierre, abierto
		FROM horarios_especiales
		WHERE fecha = $1
		ORDER BY hora_apertura;
	`, date)
	if err != nil {
		return nil, dbErr.MapSQLError(err, "ScheduleRepo.GetSpecialHoursByDate")
	}
	defer rows.Close()

	var result []models.SpecialDay
	for rows.Next() {
		var (
			id                int
			d                 time.Time
			openStr, closeStr sql.NullString
			active            bool
		)
		if err := rows.Scan(&id, &d, &openStr, &closeStr, &active); err != nil {
			return nil, appErr.Wrap("ScheduleRepo.GetSpecialHoursByDate(scan)", appErr.ErrInternal, err)
		}

		sd := models.SpecialDay{
			ID:     id,
			Date:   d,
			Active: active,
		}

		// Only build a range if active AND both times present
		if active && openStr.Valid && closeStr.Valid {
			openT, err := time.Parse("15:04:05", openStr.String)
			if err != nil {
				return nil, appErr.Wrap("ScheduleRepo.GetSpecialHoursByDate(parse open)", appErr.ErrInternal, err)
			}
			closeT, err := time.Parse("15:04:05", closeStr.String)
			if err != nil {
				return nil, appErr.Wrap("ScheduleRepo.GetSpecialHoursByDate(parse close)", appErr.ErrInternal, err)
			}
			sd.Ranges = []models.TimeRange{
				{Start: openT, End: closeT},
			}
		}

		result = append(result, sd)
	}

	// No rows found → return nil (not an error)
	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

func (r *repository) CreateSpecialHour(day models.SpecialDay) error {
	tx, err := r.db.Begin()
	if err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.CreateSpecialHour(begin)")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if day.Active && len(day.Ranges) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO horarios_especiales (fecha, hora_apertura, hora_cierre, abierto)
			VALUES ($1, $2, $3, $4);
		`)
		if err != nil {
			return dbErr.MapSQLError(err, "ScheduleRepo.CreateSpecialHour(prepare)")
		}
		defer stmt.Close()

		for _, tr := range day.Ranges {
			if !tr.IsValid() {
				return appErr.NewDomainError(appErr.ErrInvalidInput, "Rango horario inválido en CreateSpecialHour")
			}
			if _, err := stmt.Exec(day.Date, tr.Start, tr.End, true); err != nil {
				return dbErr.MapSQLError(err, "ScheduleRepo.CreateSpecialHour(insert)")
			}
		}
	} else {
		// If no ranges, insert a closed day
		if _, err := tx.Exec(`
			INSERT INTO horarios_especiales (fecha, hora_apertura, hora_cierre, abierto)
			VALUES ($1, NULL, NULL, FALSE);
		`, day.Date); err != nil {
			return dbErr.MapSQLError(err, "ScheduleRepo.CreateSpecialHour(insert closed)")
		}
	}

	if err := tx.Commit(); err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.CreateSpecialHour(commit)")
	}

	return nil
}

func (r *repository) UpdateSpecialHour(day models.SpecialDay) error {
	if day.Date.IsZero() {
		return appErr.Wrap("ScheduleRepo.UpdateSpecialHour", appErr.ErrInvalidInput, nil)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.UpdateSpecialHour(begin)")
	}
	defer func() {
		_ = tx.Rollback() // safe rollback if commit not called
	}()

	// Delete all existing entries for this date
	if _, err := tx.Exec(`DELETE FROM horarios_especiales WHERE fecha = $1;`, day.Date); err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.UpdateSpecialHour(delete)")
	}

	// Reinsert all new ranges for that date
	if day.Active && len(day.Ranges) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO horarios_especiales (fecha, hora_apertura, hora_cierre, abierto)
			VALUES ($1, $2, $3, $4);
		`)
		if err != nil {
			return dbErr.MapSQLError(err, "ScheduleRepo.UpdateSpecialHour(prepare)")
		}
		defer stmt.Close()

		for _, tr := range day.Ranges {
			if !tr.IsValid() {
				return appErr.NewDomainError(appErr.ErrInvalidInput, "Rango horario inválido en UpdateSpecialHour")
			}
			if _, err := stmt.Exec(day.Date, tr.Start, tr.End, true); err != nil {
				return dbErr.MapSQLError(err, "ScheduleRepo.UpdateSpecialHour(insert)")
			}
		}
	} else {
		// if no ranges provided, insert a closed (inactive) row
		if _, err := tx.Exec(`
			INSERT INTO horarios_especiales (fecha, hora_apertura, hora_cierre, abierto)
			VALUES ($1, NULL, NULL, FALSE);
		`, day.Date); err != nil {
			return dbErr.MapSQLError(err, "ScheduleRepo.UpdateSpecialHour(insert closed)")
		}
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.UpdateSpecialHour(commit)")
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

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

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
