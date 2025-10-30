package user

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/user/models"

	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// Repository holds all the dependencies
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new Repository instance
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetByID retrieves a user by ID
func (r *Repository) GetByID(id int) (*models.User, error) {
	query := `SELECT id, username, correo, password_hash FROM usuarios WHERE id = $1`
	var u models.User

	err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetByID: %w", appErr.ErrUserNotFound)
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}

	return &u, nil
}

// GetByUsernameOrEmail retrieves a user by username OR email
func (r *Repository) GetByUsernameOrEmail(identifier string) (*models.User, error) {
	query := `
		SELECT id, username, correo, password_hash
		FROM usuarios
		WHERE username = $1 OR correo = $1
	`
	var u models.User

	err := r.db.QueryRow(query, identifier).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetByUsernameOrEmail: %w", appErr.ErrUserNotFound)
		}
		return nil, fmt.Errorf("GetByUsernameOrEmail: %w", err)
	}

	return &u, nil
}

// Create inserts a new user into the database
func (r *Repository) Create(u *models.User) error {
	query := `INSERT INTO usuarios (username, password_hash, correo) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, u.Username, u.PasswordHash, u.Email)
	if err != nil {
		return fmt.Errorf("Create: %w", err)
	}
	return nil
}

// Delete removes a user by ID
func (r *Repository) Delete(id int) error {
	query := `DELETE FROM usuarios WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("Delete: %w", appErr.ErrUserNotFound)
	}
	return nil
}

func (r *Repository) GetRolesAndPermissions(userID int) ([]models.Role, []models.Permission, error) {
	query := `
	SELECT DISTINCT r.id, r.nombre, r.descripcion,
	                p.id, p.nombre, p.descripcion
	FROM usuarios u
	JOIN usuarios_roles ur ON ur.usuario_id = u.id
	JOIN roles r ON r.id = ur.rol_id
	LEFT JOIN roles_permisos rp ON rp.rol_id = r.id
	LEFT JOIN permisos p ON p.id = rp.permiso_id
	WHERE u.id = $1;
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("GetRolesAndPermissions: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	var perms []models.Permission
	roleSeen := make(map[int]bool)
	permSeen := make(map[int]bool)

	for rows.Next() {
		var role models.Role
		var perm models.Permission

		if err := rows.Scan(
			&role.ID, &role.Name, &role.Description,
			&perm.ID, &perm.Name, &perm.Description,
		); err != nil {
			return nil, nil, fmt.Errorf("GetRolesAndPermissions (scan): %w", err)
		}

		if role.ID != 0 && !roleSeen[role.ID] {
			roles = append(roles, role)
			roleSeen[role.ID] = true
		}
		if perm.ID != 0 && !permSeen[perm.ID] {
			perms = append(perms, perm)
			permSeen[perm.ID] = true
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("GetRolesAndPermissions (rows): %w", err)
	}

	return roles, perms, nil
}
