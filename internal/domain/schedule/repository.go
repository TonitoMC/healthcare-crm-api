//go:generate mockgen -source=repository.go -destination=./mocks/repository.go -package=mocks

package schedule

import (
	"database/sql"
	"time"

	dbErr "github.com/tonitomc/healthcare-crm-api/internal/database"
	models "github.com/tonitomc/healthcare-crm-api/internal/domain/schedule/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
	"github.com/tonitomc/healthcare-crm-api/pkg/timeutil"
)

// Repository defines the data access contract for working hours and special days.
type Repository interface {
	// Reads
	GetAllWorkingHours() ([]models.WorkDay, error)
	GetAllSpecialHours() ([]models.SpecialDay, error)
	GetSpecialHoursBetween(start, end time.Time) ([]models.SpecialDay, error)
	GetSpecialHoursByDate(date time.Time) ([]models.SpecialDay, error)

	// Writes
	UpdateWorkingHour(day models.WorkDay) error
	UpdateSpecialHour(day models.SpecialDay) error
	DeleteSpecialHour(date time.Time) error
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
				loc := timeutil.ClinicLocation()
				// Re-anclar a una fecha fija pero con timezone de clínica (solo hora interesa)
				anchoredStart := time.Date(2000, 1, 1, start.Hour(), start.Minute(), start.Second(), 0, loc)
				anchoredEnd := time.Date(2000, 1, 1, end.Hour(), end.Minute(), end.Second(), 0, loc)
				wd.Ranges = []models.TimeRange{{
					Start: anchoredStart,
					End:   anchoredEnd,
				}}
			}
		}

		result = append(result, wd)
	}

	return result, nil
}

func (r *repository) UpdateWorkingHour(day models.WorkDay) error {
	if day.DayOfWeek < 1 || day.DayOfWeek > 7 {
		return appErr.Wrap("ScheduleRepo.UpdateWorkingHour", appErr.ErrInvalidInput, nil)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.UpdateWorkingHour(begin)")
	}
	defer func() { _ = tx.Rollback() }()

	// Delete all existing entries for this weekday
	if _, err := tx.Exec(`DELETE FROM horarios_laborales WHERE dia_semana = $1;`, day.DayOfWeek); err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.UpdateWorkingHour(delete)")
	}

	// If active and has ranges → insert each one
	if day.Active && len(day.Ranges) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO horarios_laborales (dia_semana, hora_apertura, hora_cierre, abierto)
			VALUES ($1, $2, $3, TRUE);
		`)
		if err != nil {
			return dbErr.MapSQLError(err, "ScheduleRepo.UpdateWorkingHour(prepare)")
		}
		defer stmt.Close()

		for _, tr := range day.Ranges {
			if !tr.IsValid() {
				return appErr.NewDomainError(appErr.ErrInvalidInput, "Rango horario inválido en UpdateWorkingHour")
			}
			if _, err := stmt.Exec(day.DayOfWeek, tr.Start, tr.End); err != nil {
				return dbErr.MapSQLError(err, "ScheduleRepo.UpdateWorkingHour(insert range)")
			}
		}

	} else {
		// If closed or no ranges → single inactive record
		if _, err := tx.Exec(`
			INSERT INTO horarios_laborales (dia_semana, hora_apertura, hora_cierre, abierto)
			VALUES ($1, NULL, NULL, FALSE);
		`, day.DayOfWeek); err != nil {
			return dbErr.MapSQLError(err, "ScheduleRepo.UpdateWorkingHour(insert closed)")
		}
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.UpdateWorkingHour(commit)")
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
				loc := timeutil.ClinicLocation()
				anchoredStart := time.Date(2000, 1, 1, start.Hour(), start.Minute(), start.Second(), 0, loc)
				anchoredEnd := time.Date(2000, 1, 1, end.Hour(), end.Minute(), end.Second(), 0, loc)
				sd.Ranges = []models.TimeRange{{
					Start: anchoredStart,
					End:   anchoredEnd,
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
			loc := timeutil.ClinicLocation()
			anchoredStart := time.Date(2000, 1, 1, openT.Hour(), openT.Minute(), openT.Second(), 0, loc)
			anchoredEnd := time.Date(2000, 1, 1, closeT.Hour(), closeT.Minute(), closeT.Second(), 0, loc)
			sd.Ranges = []models.TimeRange{
				{Start: anchoredStart, End: anchoredEnd},
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

func (r *repository) DeleteSpecialHour(date time.Time) error {
	if date.IsZero() {
		return appErr.Wrap("ScheduleRepo.DeleteSpecialHourByDate", appErr.ErrInvalidInput, nil)
	}

	_, err := r.db.Exec(`DELETE FROM horarios_especiales WHERE fecha = $1;`, date)
	if err != nil {
		return dbErr.MapSQLError(err, "ScheduleRepo.DeleteSpecialHourByDate")
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
