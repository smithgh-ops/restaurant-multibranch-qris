package report

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
)

// Handler groups HTTP handlers for reporting endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a new report handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes mounts report endpoints under an authenticated v1 group.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	reports := v1.Group("/reports")
	reports.Use(authMiddleware)
	{
		reports.GET("/sales", h.Sales)
		reports.GET("/sales/export", h.ExportCSV)
		reports.GET("/top-items", h.TopItems)
	}
}

// parseFilters extracts shared query parameters used by every report endpoint.
// Returns orgID, optional branchID, dateFrom (YYYY-MM-DD), dateTo (YYYY-MM-DD).
// dateFrom defaults to 30 days ago; dateTo defaults to today.
func (h *Handler) parseFilters(c *gin.Context) (orgID uint64, branchID *uint64, dateFrom, dateTo string, ok bool) {
	orgID, ok = middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	if raw := c.Query("branch_id"); raw != "" {
		bid, err := strconv.ParseUint(raw, 10, 64)
		if err == nil {
			branchID = &bid
		}
	}

	now := time.Now()
	dateFrom = c.DefaultQuery("date_from", now.AddDate(0, 0, -29).Format("2006-01-02"))
	dateTo = c.DefaultQuery("date_to", now.Format("2006-01-02"))
	ok = true
	return
}

// Sales handles GET /api/v1/reports/sales.
// Query params: branch_id (optional), date_from (YYYY-MM-DD), date_to (YYYY-MM-DD).
func (h *Handler) Sales(c *gin.Context) {
	orgID, branchID, dateFrom, dateTo, ok := h.parseFilters(c)
	if !ok {
		return
	}

	summary, err := h.repo.GetSalesSummary(c.Request.Context(), orgID, branchID, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	daily, err := h.repo.GetDailySales(c.Request.Context(), orgID, branchID, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, SalesReportResponse{
		Summary: *summary,
		Daily:   daily,
	})
}

// TopItems handles GET /api/v1/reports/top-items.
// Query params: branch_id (optional), date_from, date_to, limit (default 10, max 100).
func (h *Handler) TopItems(c *gin.Context) {
	orgID, branchID, dateFrom, dateTo, ok := h.parseFilters(c)
	if !ok {
		return
	}

	limit := 10
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}

	items, err := h.repo.GetTopItems(c.Request.Context(), orgID, branchID, dateFrom, dateTo, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// ExportCSV handles GET /api/v1/reports/sales/export.
// Returns a CSV file with the daily sales data for the requested period.
func (h *Handler) ExportCSV(c *gin.Context) {
	orgID, branchID, dateFrom, dateTo, ok := h.parseFilters(c)
	if !ok {
		return
	}

	daily, err := h.repo.GetDailySales(c.Request.Context(), orgID, branchID, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	filename := "laporan_penjualan_" + dateFrom + "_" + dateTo + ".csv"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv; charset=utf-8")

	w := csv.NewWriter(c.Writer)
	// BOM for Excel compatibility
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	_ = w.Write([]string{"Tanggal", "Total Pesanan", "Pendapatan (IDR)"})
	for _, d := range daily {
		_ = w.Write([]string{d.Date, strconv.FormatInt(d.OrderCount, 10), d.Revenue})
	}
	w.Flush()
}
