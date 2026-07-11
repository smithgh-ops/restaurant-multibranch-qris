package table

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
)

// Handler groups HTTP handlers for table and dining-area endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a new table handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// ── Dining Areas ──────────────────────────────────────────────────────────────

// ListDiningAreas handles GET /api/v1/branches/:branch_id/dining-areas.
func (h *Handler) ListDiningAreas(c *gin.Context) {
	branchID, orgID, ok := h.parseBranchParam(c)
	if !ok {
		return
	}
	if !h.branchBelongsToOrg(c, branchID, orgID) {
		return
	}
	areas, err := h.repo.ListDiningAreas(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": areas})
}

// CreateDiningArea handles POST /api/v1/branches/:branch_id/dining-areas.
func (h *Handler) CreateDiningArea(c *gin.Context) {
	branchID, orgID, ok := h.parseBranchParam(c)
	if !ok {
		return
	}
	if !h.branchBelongsToOrg(c, branchID, orgID) {
		return
	}
	var req CreateDiningAreaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	area, err := h.repo.CreateDiningArea(c.Request.Context(), branchID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, area)
}

// ── Tables ────────────────────────────────────────────────────────────────────

// ListTables handles GET /api/v1/branches/:branch_id/tables.
func (h *Handler) ListTables(c *gin.Context) {
	branchID, orgID, ok := h.parseBranchParam(c)
	if !ok {
		return
	}
	if !h.branchBelongsToOrg(c, branchID, orgID) {
		return
	}
	tables, err := h.repo.ListTables(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tables})
}

// CreateTable handles POST /api/v1/branches/:branch_id/tables.
func (h *Handler) CreateTable(c *gin.Context) {
	branchID, orgID, ok := h.parseBranchParam(c)
	if !ok {
		return
	}
	if !h.branchBelongsToOrg(c, branchID, orgID) {
		return
	}
	var req CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	t, err := h.repo.CreateTable(c.Request.Context(), branchID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if isDuplicate(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "nomor meja sudah ada pada area ini"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, t)
}

// UpdateTable handles PATCH /api/v1/branches/:branch_id/tables/:id.
func (h *Handler) UpdateTable(c *gin.Context) {
	branchID, orgID, ok := h.parseBranchParam(c)
	if !ok {
		return
	}
	if !h.branchBelongsToOrg(c, branchID, orgID) {
		return
	}
	tableID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}
	var req UpdateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	t, err := h.repo.UpdateTable(c.Request.Context(), tableID, branchID, &req)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "meja tidak ditemukan"})
		return
	}
	if err != nil {
		if isDuplicate(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "nomor meja sudah ada pada area ini"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, t)
}

// RegisterRoutes mounts dining area and table endpoints.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	b := v1.Group("/branches/:branch_id", authMiddleware)
	{
		b.GET("/dining-areas", h.ListDiningAreas)
		b.POST("/dining-areas", h.CreateDiningArea)
		b.GET("/tables", h.ListTables)
		b.POST("/tables", h.CreateTable)
		b.PATCH("/tables/:id", h.UpdateTable)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (h *Handler) parseBranchParam(c *gin.Context) (branchID uint64, orgID uint64, ok bool) {
	var err error
	orgID, ok = middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	branchID, err = strconv.ParseUint(c.Param("branch_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id tidak valid"})
		ok = false
		return
	}
	return
}

func (h *Handler) branchBelongsToOrg(c *gin.Context, branchID, orgID uint64) bool {
	var foundOrgID uint64
	err := h.repo.db.QueryRowContext(c.Request.Context(),
		`SELECT organization_id FROM branches WHERE id = ? LIMIT 1`, branchID,
	).Scan(&foundOrgID)
	if err != nil || foundOrgID != orgID {
		c.JSON(http.StatusNotFound, gin.H{"error": "cabang tidak ditemukan"})
		return false
	}
	return true
}

func isDuplicate(err error) bool {
	return err != nil && strings.Contains(err.Error(), "1062")
}
