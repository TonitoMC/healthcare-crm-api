//go:generate mockgen -source=repository.go -destination=mocks/repository.go -package=mocks

package dashboard

import (
	"database/sql"
	"time"

	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/dashboard/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Repository interface {
	GetStats() (*models.DashboardStats, error)
	GetRecentActivity(limit int) ([]models.RecentActivity, error)
	GetCriticalExams(limit int) ([]models.CriticalExam, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetStats() (*models.DashboardStats, error) {
	var stats models.DashboardStats
	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Total appointments today
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM citas WHERE fecha >= $1 AND fecha < $2
	`, startOfDay, endOfDay).Scan(&stats.TotalAppointmentsToday)
	if err != nil {
		return nil, database.MapSQLError(err, "DashboardRepository.GetStats(appointments)")
	}

	// Pending exams total
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM examenes WHERE s3_key IS NULL OR s3_key = ''
	`).Scan(&stats.PendingExamsTotal)
	if err != nil {
		return nil, database.MapSQLError(err, "DashboardRepository.GetStats(exams)")
	}

	// Completed appointments (citas in the past today)
	now := time.Now()
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM citas WHERE fecha >= $1 AND fecha < $2
	`, startOfDay, now).Scan(&stats.CompletedAppointments)
	if err != nil {
		return nil, database.MapSQLError(err, "DashboardRepository.GetStats(completed)")
	}

	return &stats, nil
}

func (r *repository) GetRecentActivity(limit int) ([]models.RecentActivity, error) {
	rows, err := r.db.Query(`
		(
			SELECT 'consultation' as type,
				   'Nueva consulta: ' || co.motivo as message,
				   p.id as patient_id,
				   p.nombre as patient_name,
				   co.fecha::timestamp as timestamp
			FROM consultas co
			JOIN pacientes p ON co.paciente_id = p.id
			ORDER BY co.fecha DESC
			LIMIT $1
		)
		UNION ALL
		(
			SELECT 'exam_upload' as type,
				   'Examen subido: ' || e.tipo as message,
				   p.id as patient_id,
				   p.nombre as patient_name,
				   COALESCE(e.fecha, NOW())::timestamp as timestamp
			FROM examenes e
			JOIN pacientes p ON e.paciente_id = p.id
			WHERE e.s3_key IS NOT NULL AND e.s3_key != ''
			ORDER BY e.fecha DESC
			LIMIT $1
		)
		ORDER BY timestamp DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, database.MapSQLError(err, "DashboardRepository.GetRecentActivity")
	}
	defer rows.Close()

	var activities []models.RecentActivity
	for rows.Next() {
		var a models.RecentActivity
		if err := rows.Scan(&a.Type, &a.Message, &a.PatientID, &a.PatientName, &a.Timestamp); err != nil {
			return nil, appErr.Wrap("DashboardRepository.GetRecentActivity(scan)", appErr.ErrInternal, err)
		}
		activities = append(activities, a)
	}
	return activities, nil
}

func (r *repository) GetCriticalExams(limit int) ([]models.CriticalExam, error) {
	rows, err := r.db.Query(`
		SELECT e.id, e.paciente_id, e.consulta_id, e.tipo, e.fecha,
			   p.nombre, p.telefono,
			   CASE 
				   WHEN e.fecha IS NOT NULL AND e.fecha < CURRENT_DATE 
				   THEN EXTRACT(DAY FROM CURRENT_DATE - e.fecha)::int
				   ELSE NULL
			   END as days_overdue
		FROM examenes e
		JOIN pacientes p ON e.paciente_id = p.id
		WHERE (e.s3_key IS NULL OR e.s3_key = '')
		ORDER BY 
			CASE WHEN e.fecha IS NOT NULL THEN e.fecha ELSE CURRENT_DATE + INTERVAL '1000 days' END,
			e.id
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, database.MapSQLError(err, "DashboardRepository.GetCriticalExams")
	}
	defer rows.Close()

	var exams []models.CriticalExam
	for rows.Next() {
		var e models.CriticalExam
		if err := rows.Scan(
			&e.ID, &e.PacienteID, &e.ConsultaID, &e.Tipo, &e.Fecha,
			&e.NombrePaciente, &e.TelefonoPaciente, &e.DaysOverdue,
		); err != nil {
			return nil, appErr.Wrap("DashboardRepository.GetCriticalExams(scan)", appErr.ErrInternal, err)
		}
		exams = append(exams, e)
	}
	return exams, nil
}
