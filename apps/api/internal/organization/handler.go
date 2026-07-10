package organization

import (
	"database/sql"
	"errors"
	"net/http"

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

// RegisterRoutes mounts organization endpoints.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	v1.GET("/organization", authMiddleware, h.Get)
}
