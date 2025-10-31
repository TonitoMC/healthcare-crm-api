//go:generate mockgen -source=repository.go -destination=./mocks/repository.go -package=mocks

package user

import (
	"database/sql"

	"github.com/tonitomc/healthcare-crm-api/internal/database"
	roleModels "github.com/tonitomc/healthcare-crm-api/internal/domain/role/models"
	userModels "github.com/tonitomc/healthcare-crm-api/internal/domain/user/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// Repository defines all persistence operations for users and their roles.
type Repository interface {
	// --- User CRUD ---
	GetAll() ([]userModels.User, error)
	GetByID(id int) (*userModels.User, error)
	GetByUsernameOrEmail(identifier string) (*userModels.User, error)
	Create(u *userModels.User) error
	Update(u *userModels.User) error
	Delete(id int) error

	// --- User → Role management ---
	GetUserRoles(userID int) ([]roleModels.Role, error)
	AddRole(userID, roleID int) error
	RemoveRole(userID, roleID int) error
	ClearRoles(userID int) error
}

// Concrete implementation backed by PostgreSQL.
type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// -----------------------------------------------------------------------------
// User CRUD
// -----------------------------------------------------------------------------

func (r *repository) GetAll() ([]userModels.User, error) {
	rows, err := r.db.Query(`SELECT id, username, correo, password_hash FROM usuarios ORDER BY id`)
	if err != nil {
		return nil, database.MapSQLError(err, "UserRepository.GetAll")
	}
	defer rows.Close()

	var users []userModels.User
	for rows.Next() {
		var u userModels.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash); err != nil {
			return nil, appErr.Wrap("UserRepository.GetAll(scan)", appErr.ErrInternal, err)
		}
		users = append(users, u)
	}

	if len(users) == 0 {
		return nil, appErr.Wrap("UserRepository.GetAll", appErr.ErrNotFound, nil)
	}

	return users, nil
}

func (r *repository) GetByID(id int) (*userModels.User, error) {
	if id <= 0 {
		return nil, appErr.Wrap("UserRepository.GetByID", appErr.ErrInvalidInput, nil)
	}

	var u userModels.User
	err := r.db.QueryRow(`SELECT id, username, correo, password_hash FROM usuarios WHERE id = $1`, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash)
	if err != nil {
		return nil, database.MapSQLError(err, "UserRepository.GetByID")
	}

	return &u, nil
}

func (r *repository) GetByUsernameOrEmail(identifier string) (*userModels.User, error) {
	if identifier == "" {
		return nil, appErr.Wrap("UserRepository.GetByUsernameOrEmail", appErr.ErrInvalidInput, nil)
	}

	var u userModels.User
	err := r.db.QueryRow(`
		SELECT id, username, correo, password_hash
		FROM usuarios
		WHERE username = $1 OR correo = $1
	`, identifier).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash)
	if err != nil {
		return nil, database.MapSQLError(err, "UserRepository.GetByUsernameOrEmail")
	}

	return &u, nil
}

func (r *repository) Create(u *userModels.User) error {
	if u == nil || u.Username == "" || u.PasswordHash == "" || u.Email == "" {
		return appErr.Wrap("UserRepository.Create", appErr.ErrInvalidInput, nil)
	}

	_, err := r.db.Exec(`
		INSERT INTO usuarios (username, password_hash, correo)
		VALUES ($1, $2, $3)
	`, u.Username, u.PasswordHash, u.Email)
	if err != nil {
		return database.MapSQLError(err, "UserRepository.Create")
	}
	return nil
}

func (r *repository) Update(u *userModels.User) error {
	if u == nil || u.ID == 0 {
		return appErr.Wrap("UserRepository.Update", appErr.ErrInvalidInput, nil)
	}

	res, err := r.db.Exec(`
		UPDATE usuarios
		SET username = $1, correo = $2, password_hash = $3
		WHERE id = $4
	`, u.Username, u.Email, u.PasswordHash, u.ID)
	if err != nil {
		return database.MapSQLError(err, "UserRepository.Update")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("UserRepository.Update", appErr.ErrNotFound, nil)
	}
	return nil
}

func (r *repository) Delete(id int) error {
	if id <= 0 {
		return appErr.Wrap("UserRepository.Delete", appErr.ErrInvalidInput, nil)
	}

	res, err := r.db.Exec(`DELETE FROM usuarios WHERE id = $1`, id)
	if err != nil {
		return database.MapSQLError(err, "UserRepository.Delete")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("UserRepository.Delete", appErr.ErrNotFound, nil)
	}

	return nil
}

// -----------------------------------------------------------------------------
// User → Role management
// -----------------------------------------------------------------------------

func (r *repository) GetUserRoles(userID int) ([]roleModels.Role, error) {
	rows, err := r.db.Query(`
		SELECT r.id, r.nombre, r.descripcion
		FROM roles r
		JOIN usuarios_roles ur ON ur.rol_id = r.id
		WHERE ur.usuario_id = $1
	`, userID)
	if err != nil {
		return nil, database.MapSQLError(err, "UserRepository.GetUserRoles")
	}
	defer rows.Close()

	var roles []roleModels.Role
	for rows.Next() {
		var role roleModels.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, appErr.Wrap("UserRepository.GetUserRoles(scan)", appErr.ErrInternal, err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *repository) AddRole(userID, roleID int) error {
	if userID <= 0 || roleID <= 0 {
		return appErr.Wrap("UserRepository.AddRole", appErr.ErrInvalidInput, nil)
	}

	_, err := r.db.Exec(`INSERT INTO usuarios_roles (usuario_id, rol_id) VALUES ($1, $2)`, userID, roleID)
	if err != nil {
		return database.MapSQLError(err, "UserRepository.AddRole")
	}
	return nil
}

func (r *repository) RemoveRole(userID, roleID int) error {
	res, err := r.db.Exec(`DELETE FROM usuarios_roles WHERE usuario_id = $1 AND rol_id = $2`, userID, roleID)
	if err != nil {
		return database.MapSQLError(err, "UserRepository.RemoveRole")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("UserRepository.RemoveRole", appErr.ErrNotFound, nil)
	}

	return nil
}

func (r *repository) ClearRoles(userID int) error {
	_, err := r.db.Exec(`DELETE FROM usuarios_roles WHERE usuario_id = $1`, userID)
	if err != nil {
		return database.MapSQLError(err, "UserRepository.ClearRoles")
	}
	return nil
}
