package models

type Diagnostic struct {
	ID            int
	ConsultaID    int
	Nombre        string
	Recomendacion *string
}

// DiagnosticWithTreatments nests treatments under a diagnostic.
type DiagnosticWithTreatments struct {
	ID            int         `json:"id"`
	ConsultaID    int         `json:"consulta_id"`
	Nombre        string      `json:"nombre"`
	Recomendacion *string     `json:"recomendacion,omitempty"`
	Treatments    []Treatment `json:"treatments"`
}
