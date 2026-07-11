package selforder

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
)

// Handler groups HTTP handlers for self-order and QR token endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a selforder handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// ── Protected endpoints (JWT required) ───────────────────────────────────────

// GenerateToken handles POST /api/v1/branches/:branch_id/tables/:table_id/qr-token.
func (h *Handler) GenerateToken(c *gin.Context) {
	branchID, orgID, ok := parseBranchAndOrg(c)
	if !ok {
		return
	}
	tableID, err := strconv.ParseUint(c.Param("table_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table_id tidak valid"})
		return
	}
	token, err := h.repo.GenerateToken(c.Request.Context(), tableID, branchID, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "meja atau cabang tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, token)
}

// GetToken handles GET /api/v1/branches/:branch_id/tables/:table_id/qr-token.
func (h *Handler) GetToken(c *gin.Context) {
	branchID, orgID, ok := parseBranchAndOrg(c)
	if !ok {
		return
	}
	tableID, err := strconv.ParseUint(c.Param("table_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table_id tidak valid"})
		return
	}
	token, err := h.repo.GetToken(c.Request.Context(), tableID, branchID, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "token QR belum dibuat untuk meja ini"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, token)
}

// ── Public endpoints (no auth) ────────────────────────────────────────────────

// GetMenu handles GET /api/v1/public/table/:token.
func (h *Handler) GetMenu(c *gin.Context) {
	token := c.Param("token")
	if len(token) != 64 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token tidak valid"})
		return
	}
	menu, err := h.repo.GetMenuByToken(c.Request.Context(), token)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "token QR tidak valid atau sudah tidak aktif"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, menu)
}

// CreateOrder handles POST /api/v1/public/table/:token/orders.
func (h *Handler) CreateOrder(c *gin.Context) {
	token := c.Param("token")
	if len(token) != 64 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token tidak valid"})
		return
	}
	var req CreateSelfOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	order, err := h.repo.CreateOrderByToken(c.Request.Context(), token, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

// RegisterRoutes mounts self-order endpoints on the router.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	// Protected: token management
	protected := v1.Group("/branches/:branch_id/tables/:table_id/qr-token", authMiddleware)
	{
		protected.GET("", h.GetToken)
		protected.POST("", h.GenerateToken)
	}

	// Public: self-order flow (no auth)
	pub := v1.Group("/public/table/:token")
	{
		pub.GET("", h.GetMenu)
		pub.POST("/orders", h.CreateOrder)
	}
}

func parseBranchAndOrg(c *gin.Context) (branchID uint64, orgID uint64, ok bool) {
	orgID, ok = middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	branchID, err := strconv.ParseUint(c.Param("branch_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id tidak valid"})
		ok = false
		return
	}
	return
}
