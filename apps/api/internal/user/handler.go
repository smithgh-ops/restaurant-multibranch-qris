package user

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
)

// Handler groups HTTP handlers for user management endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a new user handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// requireOrgAdmin checks that the caller has org_admin or super_admin role. Returns false and writes
// the error response if the check fails.
func (h *Handler) requireOrgAdmin(c *gin.Context, userID, orgID uint64) bool {
	ok, err := h.repo.IsOrgAdmin(c.Request.Context(), userID, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return false
	}
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "akses ditolak"})
		return false
	}
	return true
}

// ListRoles handles GET /api/v1/roles — returns all available roles.
func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.repo.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

// List handles GET /api/v1/users.
func (h *Handler) List(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	var branchID *uint64
	if raw := c.Query("branch_id"); raw != "" {
		v, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id tidak valid"})
			return
		}
		branchID = &v
	}

	users, err := h.repo.List(c.Request.Context(), orgID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// Get handles GET /api/v1/users/:id.
func (h *Handler) Get(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	u, err := h.repo.FindByID(c.Request.Context(), id, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "pengguna tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// Create handles POST /api/v1/users — requires org_admin.
func (h *Handler) Create(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	callerID, ok2 := middleware.UserIDFromContext(c)
	if !ok || !ok2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	if !h.requireOrgAdmin(c, callerID, orgID) {
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	u, err := h.repo.Create(c.Request.Context(), orgID, &req, string(hash))
	if err != nil {
		if isDuplicateEntry(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "email sudah digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, u)
}

// Update handles PATCH /api/v1/users/:id — requires org_admin.
func (h *Handler) Update(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	callerID, ok2 := middleware.UserIDFromContext(c)
	if !ok || !ok2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	if !h.requireOrgAdmin(c, callerID, orgID) {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}

	u, err := h.repo.Update(c.Request.Context(), id, orgID, &req)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "pengguna tidak ditemukan"})
		return
	}
	if err != nil {
		if isDuplicateEntry(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "email sudah digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// Delete handles DELETE /api/v1/users/:id — deactivates a user; requires org_admin.
func (h *Handler) Delete(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	callerID, ok2 := middleware.UserIDFromContext(c)
	if !ok || !ok2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	if !h.requireOrgAdmin(c, callerID, orgID) {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	if err := h.repo.Deactivate(c.Request.Context(), id, orgID); errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "pengguna tidak ditemukan"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pengguna dinonaktifkan"})
}

// UpdateRoles handles PUT /api/v1/users/:id/roles — replaces all roles; requires org_admin.
func (h *Handler) UpdateRoles(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	callerID, ok2 := middleware.UserIDFromContext(c)
	if !ok || !ok2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	if !h.requireOrgAdmin(c, callerID, orgID) {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	var req UpdateRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}

	u, err := h.repo.SetRoles(c.Request.Context(), id, orgID, req.Roles)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "pengguna tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// RegisterRoutes mounts user management endpoints on the v1 group.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	v1.GET("/roles", authMiddleware, h.ListRoles)

	users := v1.Group("/users", authMiddleware)
	{
		users.GET("", h.List)
		users.POST("", h.Create)
		users.GET("/:id", h.Get)
		users.PATCH("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
		users.PUT("/:id/roles", h.UpdateRoles)
	}
}

// isDuplicateEntry checks if the error is a MySQL duplicate key error.
func isDuplicateEntry(err error) bool {
	return err != nil && strings.Contains(err.Error(), "1062")
}
