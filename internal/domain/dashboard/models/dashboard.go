package models

import "time"

type DashboardStats struct {
	TotalAppointmentsToday int `json:"total_appointments_today"`
	PendingExamsTotal      int `json:"pending_exams_total"`
	CompletedAppointments  int `json:"completed_appointments"`
}

type RecentActivity struct {
	Type        string    `json:"type"` // "consultation", "exam_upload", "medical_record_update"
	Message     string    `json:"message"`
	PatientID   int       `json:"patient_id"`
	PatientName string    `json:"patient_name"`
	Timestamp   time.Time `json:"timestamp"`
}

type CriticalExam struct {
	ID               int        `json:"id"`
	PacienteID       int        `json:"paciente_id"`
	ConsultaID       *int       `json:"consulta_id,omitempty"`
	Tipo             string     `json:"tipo"`
	Fecha            *time.Time `json:"fecha,omitempty"`
	NombrePaciente   string     `json:"nombre_paciente"`
	TelefonoPaciente string     `json:"telefono_paciente"`
	DaysOverdue      *int       `json:"days_overdue,omitempty"` // null if not overdue
}
