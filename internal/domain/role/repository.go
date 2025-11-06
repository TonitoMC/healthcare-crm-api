//go:generate mockgen -source=repository.go -destination=mocks/repository.go -package=mocks

package role

import (
	"database/sql"

	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/role/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// Repository defines all operations related to roles and permissions.
type Repository interface {
	// Role CRUD
	GetAll() ([]models.Role, error)
	GetByID(id int) (*models.Role, error)
	Create(role *models.Role) error
	Update(role *models.Role) error
	Delete(id int) error

	// Permissions for roles
	GetAllPermissions() ([]models.Permission, error)
	GetPermissions(roleID int) ([]models.Permission, error)
	AddPermission(roleID, permissionID int) error
	RemovePermission(roleID, permissionID int) error
	ClearPermissions(roleID int) error
}

// repository is the concrete implementation using *sql.DB.
type repository struct {
	db *sql.DB
}

// NewRepository constructs a role repository.
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

//
// --- Role CRUD ---
//

// GetAll retrieves all roles from the database.
func (r *repository) GetAll() ([]models.Role, error) {
	rows, err := r.db.Query(`SELECT id, nombre, descripcion FROM roles ORDER BY id`)
	if err != nil {
		return nil, database.MapSQLError(err, "RoleRepository.GetAll")
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, appErr.Wrap("RoleRepository.GetAll(scan)", appErr.ErrInternal, err)
		}
		roles = append(roles, role)
	}

	if len(roles) == 0 {
		return nil, appErr.Wrap("RoleRepository.GetAll", appErr.ErrNotFound, nil)
	}

	return roles, nil
}

// GetByID retrieves a specific role by ID.
func (r *repository) GetByID(id int) (*models.Role, error) {
	var role models.Role
	err := r.db.QueryRow(
		`SELECT id, nombre, descripcion FROM roles WHERE id = $1`,
		id,
	).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		return nil, database.MapSQLError(err, "RoleRepository.GetByID")
	}
	return &role, nil
}

// Create inserts a new role into the database.
func (r *repository) Create(role *models.Role) error {
	if role == nil || role.Name == "" || role.Description == "" {
		return appErr.Wrap("RoleRepository.Create", appErr.ErrInvalidInput, nil)
	}
	_, err := r.db.Exec(
		`INSERT INTO roles (nombre, descripcion) VALUES ($1, $2)`,
		role.Name, role.Description,
	)
	if err != nil {
		return database.MapSQLError(err, "RoleRepository.Create")
	}
	return nil
}

// Update modifies an existing role.
func (r *repository) Update(role *models.Role) error {
	if role == nil || role.ID == 0 || role.Description == "" {
		return appErr.Wrap("RoleRepository.Update", appErr.ErrInvalidInput, nil)
	}

	res, err := r.db.Exec(
		`UPDATE roles SET nombre = $1, descripcion = $2 WHERE id = $3`,
		role.Name, role.Description, role.ID,
	)
	if err != nil {
		return database.MapSQLError(err, "RoleRepository.Update")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("RoleRepository.Update", appErr.ErrNotFound, nil)
	}

	return nil
}

// Delete removes a role by ID.
func (r *repository) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM roles WHERE id = $1`, id)
	if err != nil {
		return database.MapSQLError(err, "RoleRepository.Delete")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("RoleRepository.Delete", appErr.ErrNotFound, nil)
	}

	return nil
}

//
// --- Role → Permission management ---
//

// GetPermissions retrieves all permissions for a given role.
func (r *repository) GetPermissions(roleID int) ([]models.Permission, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.nombre, p.descripcion
		FROM permisos p
		JOIN roles_permisos rp ON rp.permiso_id = p.id
		WHERE rp.rol_id = $1`, roleID)
	if err != nil {
		return nil, database.MapSQLError(err, "RoleRepository.GetPermissions")
	}
	defer rows.Close()

	var perms []models.Permission
	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			return nil, appErr.Wrap("RoleRepository.GetPermissions(scan)", appErr.ErrInternal, err)
		}
		perms = append(perms, p)
	}

	if len(perms) == 0 {
		return nil, appErr.Wrap("RoleRepository.GetPermissions", appErr.ErrNotFound, nil)
	}

	return perms, nil
}

// AddPermission links a permission to a role.
func (r *repository) AddPermission(roleID, permissionID int) error {
	_, err := r.db.Exec(
		`INSERT INTO roles_permisos (rol_id, permiso_id) VALUES ($1, $2)`,
		roleID, permissionID,
	)
	if err != nil {
		return database.MapSQLError(err, "RoleRepository.AddPermission")
	}
	return nil
}

// RemovePermission unlinks a permission from a role.
func (r *repository) RemovePermission(roleID, permissionID int) error {
	res, err := r.db.Exec(
		`DELETE FROM roles_permisos WHERE rol_id = $1 AND permiso_id = $2`,
		roleID, permissionID,
	)
	if err != nil {
		return database.MapSQLError(err, "RoleRepository.RemovePermission")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("RoleRepository.RemovePermission", appErr.ErrNotFound, nil)
	}

	return nil
}

// ClearPermissions removes all permissions from a given role.
func (r *repository) ClearPermissions(roleID int) error {
	_, err := r.db.Exec(`DELETE FROM roles_permisos WHERE rol_id = $1`, roleID)
	if err != nil {
		return database.MapSQLError(err, "RoleRepository.ClearPermissions")
	}
	return nil
}

// GetAllPermissions retrieves all permissions in the system.
func (r *repository) GetAllPermissions() ([]models.Permission, error) {
	rows, err := r.db.Query(`SELECT id, nombre, descripcion FROM permisos ORDER BY id`)
	if err != nil {
		return nil, database.MapSQLError(err, "RoleRepository.GetAllPermissions")
	}
	defer rows.Close()

	var perms []models.Permission
	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			return nil, appErr.Wrap("RoleRepository.GetAllPermissions(scan)", appErr.ErrInternal, err)
		}
		perms = append(perms, p)
	}

	if len(perms) == 0 {
		return nil, appErr.Wrap("RoleRepository.GetAllPermissions", appErr.ErrNotFound, nil)
	}

	return perms, nil
}
