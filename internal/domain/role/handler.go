package role

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	roleModels "github.com/tonitomc/healthcare-crm-api/internal/domain/role/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// Handler exposes HTTP endpoints for role operations.
type Handler struct {
	service Service
}

// NewHandler constructs a new RoleHandler.
func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

// RegisterRoutes mounts /role routes under the provided Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group) {
	roleGroup := g.Group("/role", ErrorMiddleware())

	roleGroup.GET("/all/permissions", h.GetAllPermissions, middleware.RequirePermission("manejar-roles"))

	// --- Role CRUD ---
	roleGroup.GET("", h.GetAllRoles, middleware.RequirePermission("manejar-roles"))
	roleGroup.GET("/:id", h.GetRoleByID, middleware.RequirePermission("manejar-roles"))
	roleGroup.POST("", h.CreateRole, middleware.RequirePermission("manejar-roles"))
	roleGroup.PUT("/:id", h.UpdateRole, middleware.RequirePermission("manejar-roles"))
	roleGroup.DELETE("/:id", h.DeleteRole, middleware.RequirePermission("manejar-roles"))

	// --- Permissions ---
	roleGroup.GET("/:id/permissions", h.GetPermissions, middleware.RequirePermission("manejar-roles"))
	roleGroup.POST("/:id/permissions", h.AddPermission, middleware.RequirePermission("manejar-roles"))
	roleGroup.DELETE("/:id/permissions/:permissionID", h.RemovePermission, middleware.RequirePermission("manejar-roles"))
	roleGroup.PUT("/:id/permissions", h.UpdateRolePermissions, middleware.RequirePermission("manejar-roles"))
}

// -----------------------------------------------------------------------------
// Role CRUD
// -----------------------------------------------------------------------------

// GET /role
func (h *Handler) GetAllRoles(c echo.Context) error {
	roles, err := h.service.GetAllRoles()
	if err != nil {
		return err
	}

	if len(roles) == 0 {
		return c.JSON(http.StatusOK, echo.Map{
			"message": "No hay roles registrados",
			"data":    []roleModels.Role{},
		})
	}

	return c.JSON(http.StatusOK, roles)
}

// GET /role/:id
func (h *Handler) GetRoleByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("RoleHandler.GetRoleByID.ParseID", appErr.ErrInvalidInput, err)
	}

	role, perms, err := h.service.GetRoleByID(id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"role":        role,
		"permissions": perms,
	})
}

// POST /role
func (h *Handler) CreateRole(c echo.Context) error {
	var req roleModels.Role
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("RoleHandler.CreateRole.Bind", appErr.ErrInvalidInput, err)
	}

	if err := h.service.CreateRole(&req); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "Rol creado correctamente"})
}

// PUT /role/:id
func (h *Handler) UpdateRole(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("RoleHandler.UpdateRole.ParseID", appErr.ErrInvalidInput, err)
	}

	var req roleModels.Role
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("RoleHandler.UpdateRole.Bind", appErr.ErrInvalidInput, err)
	}
	req.ID = id

	if err := h.service.UpdateRole(&req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Rol actualizado correctamente"})
}

// DELETE /role/:id
func (h *Handler) DeleteRole(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("RoleHandler.DeleteRole.ParseID", appErr.ErrInvalidInput, err)
	}

	if err := h.service.DeleteRole(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Rol eliminado correctamente"})
}

// -----------------------------------------------------------------------------
// Permissions
// -----------------------------------------------------------------------------

// GET /role/:id/permissions
func (h *Handler) GetPermissions(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("RoleHandler.GetPermissions.ParseID", appErr.ErrInvalidInput, err)
	}

	perms, err := h.service.GetPermissions(id)
	if err != nil {
		return err
	}

	if len(perms) == 0 {
		return c.JSON(http.StatusOK, echo.Map{
			"message": "Este rol no tiene permisos asignados",
			"data":    []roleModels.Permission{},
		})
	}

	return c.JSON(http.StatusOK, perms)
}

// POST /role/:id/permissions
func (h *Handler) AddPermission(c echo.Context) error {
	roleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("RoleHandler.AddPermission.ParseID", appErr.ErrInvalidInput, err)
	}

	var payload struct {
		PermissionID int `json:"permission_id"`
	}
	if err := c.Bind(&payload); err != nil {
		return appErr.Wrap("RoleHandler.AddPermission.Bind", appErr.ErrInvalidInput, err)
	}

	if err := h.service.AddPermission(roleID, payload.PermissionID); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "Permiso asignado correctamente"})
}

// DELETE /role/:id/permissions/:permissionID
func (h *Handler) RemovePermission(c echo.Context) error {
	roleID, err1 := strconv.Atoi(c.Param("id"))
	permID, err2 := strconv.Atoi(c.Param("permissionID"))
	if err1 != nil || err2 != nil {
		return appErr.Wrap("RoleHandler.RemovePermission.ParseIDs", appErr.ErrInvalidInput, nil)
	}

	if err := h.service.RemovePermission(roleID, permID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Permiso eliminado correctamente"})
}

// PUT /role/:id/permissions
func (h *Handler) UpdateRolePermissions(c echo.Context) error {
	roleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("RoleHandler.UpdateRolePermissions.ParseID", appErr.ErrInvalidInput, err)
	}

	var payload struct {
		PermissionIDs []int `json:"permission_ids"`
	}
	if err := c.Bind(&payload); err != nil {
		return appErr.Wrap("RoleHandler.UpdateRolePermissions.Bind", appErr.ErrInvalidInput, err)
	}

	if err := h.service.UpdateRolePermissions(roleID, payload.PermissionIDs); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Permisos actualizados correctamente"})
}

// GET /role/permissions
func (h *Handler) GetAllPermissions(c echo.Context) error {
	perms, err := h.service.GetAllPermissions()
	if err != nil {
		return err
	}

	if len(perms) == 0 {
		return c.JSON(http.StatusOK, echo.Map{
			"message": "No hay permisos registrados",
			"data":    []roleModels.Permission{},
		})
	}

	return c.JSON(http.StatusOK, perms)
}
