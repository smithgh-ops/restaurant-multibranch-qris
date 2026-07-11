package payment

import (
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
)

// Handler groups HTTP handlers for payment endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a payment handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// GetGatewayConfig handles GET /api/v1/branches/:branch_id/payment-gateway-config.
func (h *Handler) GetGatewayConfig(c *gin.Context) {
	branchID, orgID, ok := parseBranchAndOrg(c)
	if !ok {
		return
	}
	cfg, err := h.repo.LoadGatewayConfig(c.Request.Context(), branchID, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "konfigurasi gateway tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	cfg.Config["api_key"] = ""
	cfg.Config["webhook_secret"] = ""
	c.JSON(http.StatusOK, cfg)
}

// UpsertGatewayConfig handles PUT /api/v1/branches/:branch_id/payment-gateway-config.
func (h *Handler) UpsertGatewayConfig(c *gin.Context) {
	branchID, orgID, ok := parseBranchAndOrg(c)
	if !ok {
		return
	}
	var req UpsertGatewayConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	cfg, err := h.repo.UpsertGatewayConfig(c.Request.Context(), branchID, orgID, &req)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "cabang tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg.Config["api_key"] = ""
	cfg.Config["webhook_secret"] = ""
	c.JSON(http.StatusOK, cfg)
}

// CreateInvoice handles POST /api/v1/orders/:id/payments/qris-invoice.
func (h *Handler) CreateInvoice(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}
	var req CreateInvoiceRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
			return
		}
	}
	p, err := h.repo.CreateInvoice(c.Request.Context(), orderID, orgID, &req)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "pesanan tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// ListReconciliation handles GET /api/v1/payments/reconciliation.
func (h *Handler) ListReconciliation(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	branchID := parseUintPointer(c.Query("branch_id"))
	var status *string
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		if _, valid := validPaymentStatuses[raw]; !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status tidak valid"})
			return
		}
		status = &raw
	}

	var dateFrom *time.Time
	if raw := strings.TrimSpace(c.Query("date_from")); raw != "" {
		if parsed, err := time.Parse("2006-01-02", raw); err == nil {
			dateFrom = &parsed
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date_from tidak valid (YYYY-MM-DD)"})
			return
		}
	}
	var dateTo *time.Time
	if raw := strings.TrimSpace(c.Query("date_to")); raw != "" {
		if parsed, err := time.Parse("2006-01-02", raw); err == nil {
			end := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			dateTo = &end
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date_to tidak valid (YYYY-MM-DD)"})
			return
		}
	}

	rows, err := h.repo.ListReconciliation(c.Request.Context(), orgID, branchID, status, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// Webhook handles POST /api/v1/payments/webhooks/qris.
func (h *Handler) Webhook(c *gin.Context) {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload webhook tidak valid"})
		return
	}

	// Extract branch_id from query param so we can load the right gateway config.
	branchID, err := strconv.ParseUint(strings.TrimSpace(c.Query("branch_id")), 10, 64)
	if err != nil || branchID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter branch_id wajib ada dan valid"})
		return
	}

	signature := c.GetHeader("X-QRIS-Signature")
	event, err := h.repo.NormalizeWebhook(c.Request.Context(), branchID, raw, signature)
	if err != nil {
		slog.Warn("webhook rejected", "branch_id", branchID, "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if _, ok := validPaymentStatuses[event.Status]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status pembayaran tidak valid"})
		return
	}

	p, err := h.repo.ProcessWebhook(c.Request.Context(), branchID, event, raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "webhook diproses", "data": p})
}

// RegisterRoutes mounts payment endpoints.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	branches := v1.Group("/branches/:branch_id/payment-gateway-config", authMiddleware)
	{
		branches.GET("", h.GetGatewayConfig)
		branches.PUT("", h.UpsertGatewayConfig)
	}

	orders := v1.Group("/orders", authMiddleware)
	{
		orders.POST("/:id/payments/qris-invoice", h.CreateInvoice)
	}

	payments := v1.Group("/payments")
	{
		payments.GET("/reconciliation", authMiddleware, h.ListReconciliation)
		payments.POST("/webhooks/qris", h.Webhook)
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
