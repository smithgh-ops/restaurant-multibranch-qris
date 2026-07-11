package organization

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
)

// Handler groups HTTP handlers for the organization endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a new organization handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// Get handles GET /api/v1/organization.
// Returns the organization context of the authenticated user.
func (h *Handler) Get(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	org, err := h.repo.FindByID(c.Request.Context(), orgID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "organisasi tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, org)
}

// Update handles PATCH /api/v1/organization — updates name/slug.
func (h *Handler) Update(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	var body struct {
		Name *string `json:"name"`
		Slug *string `json:"slug"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}

	org, err := h.repo.Update(c.Request.Context(), orgID, body.Name, body.Slug)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "organisasi tidak ditemukan"})
		return
	}
	if err != nil {
		if strings.Contains(err.Error(), "1062") {
			c.JSON(http.StatusConflict, gin.H{"error": "slug sudah digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, org)
}

// RegisterRoutes mounts organization endpoints.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	v1.GET("/organization", authMiddleware, h.Get)
	v1.PATCH("/organization", authMiddleware, h.Update)
}
