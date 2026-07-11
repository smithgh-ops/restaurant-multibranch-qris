package branch

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
)

// Handler groups HTTP handlers for the branch endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a new branch handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// List handles GET /api/v1/branches.
func (h *Handler) List(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	activeFilter, err := parseOptionalActiveParam(c.Query("active"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameter active tidak valid"})
		return
	}

	branches, err := h.repo.List(c.Request.Context(), orgID, activeFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": branches})
}

// Get handles GET /api/v1/branches/:branch_id.
func (h *Handler) Get(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	id, err := strconv.ParseUint(c.Param("branch_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id tidak valid"})
		return
	}

	b, err := h.repo.FindByID(c.Request.Context(), id, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "cabang tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, b)
}

// Create handles POST /api/v1/branches.
func (h *Handler) Create(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	var req CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}

	b, err := h.repo.Create(c.Request.Context(), orgID, &req)
	if err != nil {
		if isDuplicateEntry(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "slug sudah digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, b)
}

// Update handles PATCH /api/v1/branches/:branch_id.
func (h *Handler) Update(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	id, err := strconv.ParseUint(c.Param("branch_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id tidak valid"})
		return
	}

	var req UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}

	b, err := h.repo.Update(c.Request.Context(), id, orgID, &req)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "cabang tidak ditemukan"})
		return
	}
	if err != nil {
		if isDuplicateEntry(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "slug sudah digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, b)
}

// RegisterRoutes mounts branch endpoints.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	branches := v1.Group("/branches", authMiddleware)
	{
		branches.GET("", h.List)
		branches.POST("", h.Create)
		branches.GET("/:branch_id", h.Get)
		branches.PATCH("/:branch_id", h.Update)
	}
}

// isDuplicateEntry checks if the error is a MySQL duplicate key error.
func isDuplicateEntry(err error) bool {
	return err != nil && strings.Contains(err.Error(), "1062")
}

func parseOptionalActiveParam(raw string) (*bool, error) {
	if raw == "" {
		return nil, nil
	}
	active, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, err
	}
	return &active, nil
}
