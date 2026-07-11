package order

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
)

// Handler groups HTTP handlers for order endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a new order handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// List handles GET /api/v1/orders.
// Query params: branch_id (optional), status (optional).
func (h *Handler) List(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	var branchID *uint64
	if raw := c.Query("branch_id"); raw != "" {
		bid, err := strconv.ParseUint(raw, 10, 64)
		if err == nil {
			branchID = &bid
		}
	}

	var status *string
	if raw := c.Query("status"); raw != "" {
		if _, valid := validStatuses[raw]; valid {
			status = &raw
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status tidak valid"})
			return
		}
	}

	orders, err := h.repo.List(c.Request.Context(), orgID, branchID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": orders})
}

// Get handles GET /api/v1/orders/:id.
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
	o, err := h.repo.FindByID(c.Request.Context(), id, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "pesanan tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, o)
}

// Create handles POST /api/v1/orders.
func (h *Handler) Create(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}

	userID, _ := middleware.UserIDFromContext(c)
	var createdBy *uint64
	if userID != 0 {
		createdBy = &userID
	}

	o, err := h.repo.Create(c.Request.Context(), orgID, &req, createdBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, o)
}

// UpdateStatus handles PATCH /api/v1/orders/:id/status.
func (h *Handler) UpdateStatus(c *gin.Context) {
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
	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	if _, valid := validStatuses[req.Status]; !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status tidak valid"})
		return
	}

	o, err := h.repo.UpdateStatus(c.Request.Context(), id, orgID, req.Status)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "pesanan tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, o)
}

// RegisterRoutes mounts order endpoints.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	orders := v1.Group("/orders", authMiddleware)
	{
		orders.GET("", h.List)
		orders.POST("", h.Create)
		orders.GET("/:id", h.Get)
		orders.PATCH("/:id/status", h.UpdateStatus)
	}
}
