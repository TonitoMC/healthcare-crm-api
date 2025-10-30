// internal/domain/role/repository.go
package role

import (
	"database/sql"
	"fmt"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/user/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByID(id int) (*models.Role, error) {
	var role models.Role
	query := `SELECT id, nombre, descripcion FROM roles WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}

	return &role, nil
}
