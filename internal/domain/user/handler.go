package user

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	userModels "github.com/tonitomc/healthcare-crm-api/internal/domain/user/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// Handler exposes HTTP endpoints for user operations.
type Handler struct {
	service Service
}

// NewHandler constructs a new UserHandler.
func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

// RegisterRoutes mounts /user routes under the provided Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group) {
	userGroup := g.Group("/user", ErrorMiddleware())

	// Read operations
	userGroup.GET("", h.GetAll, middleware.RequirePermission("manejar-usuarios"))
	userGroup.GET("/:id", h.GetByID, middleware.RequirePermission("manejar-usuarios"))
	userGroup.GET("/search", h.GetByUsernameOrEmail, middleware.RequirePermission("manejar-usuarios"))
	userGroup.GET("/:id/roles", h.GetUserRoles, middleware.RequirePermission("manejar-usuarios"))
	userGroup.GET("/:id/roles-permissions", h.GetRolesAndPermissions, middleware.RequirePermission("manejar-usuarios"))

	// Write operations
	userGroup.PUT("/:id", h.UpdateUser, middleware.RequirePermission("manejar-usuarios"))
	userGroup.DELETE("/:id", h.DeleteUser, middleware.RequirePermission("manejar-usuarios"))
	userGroup.POST("/:id/roles/:roleID", h.AddRole, middleware.RequirePermission("manejar-usuarios"))
	userGroup.DELETE("/:id/roles/:roleID", h.RemoveRole, middleware.RequirePermission("manejar-usuarios"))
	userGroup.DELETE("/:id/roles", h.ClearRoles, middleware.RequirePermission("manejar-usuarios"))

	userGroup.GET("/enriched", h.GetAllWithRoles, middleware.RequirePermission("manejar-usuarios"))
}

// -----------------------------------------------------------------------------
// Handlers
// -----------------------------------------------------------------------------

// GET /user
func (h *Handler) GetAll(c echo.Context) error {
	users, err := h.service.GetAllUsers()
	if err != nil {
		return err
	}

	if len(users) == 0 {
		return c.JSON(http.StatusOK, echo.Map{
			"message": "No hay usuarios registrados",
			"data":    []userModels.User{},
		})
	}

	return c.JSON(http.StatusOK, users)
}

// GET /user/:id
func (h *Handler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("UserHandler.GetByID.ParseID", appErr.ErrInvalidInput, err)
	}

	u, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, u)
}

// GET /user/search?identifier=username_or_email
func (h *Handler) GetByUsernameOrEmail(c echo.Context) error {
	identifier := c.QueryParam("identifier")
	if identifier == "" {
		return appErr.Wrap("UserHandler.GetByUsernameOrEmail", appErr.ErrInvalidInput, nil)
	}

	u, err := h.service.GetByUsernameOrEmail(identifier)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, u)
}

// PUT /user/:id
func (h *Handler) UpdateUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("UserHandler.UpdateUser.ParseID", appErr.ErrInvalidInput, err)
	}

	var req userModels.User
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("UserHandler.UpdateUser.Bind", appErr.ErrInvalidInput, err)
	}
	req.ID = id

	if err := h.service.UpdateUser(&req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Usuario actualizado correctamente"})
}

// DELETE /user/:id
func (h *Handler) DeleteUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("UserHandler.DeleteUser.ParseID", appErr.ErrInvalidInput, err)
	}

	if err := h.service.DeleteUser(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Usuario eliminado correctamente"})
}

// GET /user/:id/roles
func (h *Handler) GetUserRoles(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("UserHandler.GetUserRoles.ParseID", appErr.ErrInvalidInput, err)
	}

	roles, err := h.service.GetUserRoles(userID)
	if err != nil {
		return err
	}

	if len(roles) == 0 {
		return c.JSON(http.StatusOK, "Este usuario no tiene roles")
	}

	return c.JSON(http.StatusOK, roles)
}

// GET /user/:id/roles-permissions
func (h *Handler) GetRolesAndPermissions(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("UserHandler.GetRolesAndPermissions.ParseID", appErr.ErrInvalidInput, err)
	}

	roles, perms, err := h.service.GetRolesAndPermissions(userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"roles":       roles,
		"permissions": perms,
	})
}

// POST /user/:id/roles/:roleID
func (h *Handler) AddRole(c echo.Context) error {
	userID, err1 := strconv.Atoi(c.Param("id"))
	roleID, err2 := strconv.Atoi(c.Param("roleID"))
	if err1 != nil || err2 != nil {
		return appErr.Wrap("UserHandler.AddRole.ParseIDs", appErr.ErrInvalidInput, nil)
	}

	if err := h.service.AddRole(userID, roleID); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "Rol asignado correctamente"})
}

// DELETE /user/:id/roles/:roleID
func (h *Handler) RemoveRole(c echo.Context) error {
	userID, err1 := strconv.Atoi(c.Param("id"))
	roleID, err2 := strconv.Atoi(c.Param("roleID"))
	if err1 != nil || err2 != nil {
		return appErr.Wrap("UserHandler.RemoveRole.ParseIDs", appErr.ErrInvalidInput, nil)
	}

	if err := h.service.RemoveRole(userID, roleID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Rol eliminado correctamente"})
}

// DELETE /user/:id/roles
func (h *Handler) ClearRoles(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("UserHandler.ClearRoles.ParseID", appErr.ErrInvalidInput, err)
	}

	if err := h.service.ClearRoles(userID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Roles del usuario eliminados correctamente"})
}

func (h *Handler) GetAllWithRoles(c echo.Context) error {
	users, err := h.service.GetAllUsers()
	if err != nil {
		return err
	}

	enriched := make([]map[string]interface{}, 0, len(users))

	for _, u := range users {
		roles, err := h.service.GetUserRoles(u.ID)
		if err != nil {
			return err
		}

		readable := []string{}
		for _, r := range roles {
			readable = append(readable, r.Name) // or r.Nombre depending on model
		}

		enriched = append(enriched, map[string]interface{}{
			"id":            u.ID,
			"username":      u.Username,
			"correo":        u.Email,
			"roles":         roles,
			"rolesReadable": readable,
		})
	}

	return c.JSON(http.StatusOK, enriched)
}
