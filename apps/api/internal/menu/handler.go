package menu

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
)

// Handler groups HTTP handlers for menu endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a new menu handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// ── Categories ────────────────────────────────────────────────────────────────

// ListCategories handles GET /api/v1/menu/categories.
func (h *Handler) ListCategories(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	cats, err := h.repo.ListCategories(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cats})
}

// CreateCategory handles POST /api/v1/menu/categories.
func (h *Handler) CreateCategory(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	cat, err := h.repo.CreateCategory(c.Request.Context(), orgID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

// UpdateCategory handles PATCH /api/v1/menu/categories/:id.
func (h *Handler) UpdateCategory(c *gin.Context) {
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
	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	cat, err := h.repo.UpdateCategory(c.Request.Context(), id, orgID, &req)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "kategori tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, cat)
}

// ── Items ─────────────────────────────────────────────────────────────────────

// ListItems handles GET /api/v1/menu/items.
func (h *Handler) ListItems(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	var categoryID *uint64
	if raw := c.Query("category_id"); raw != "" {
		cid, err := strconv.ParseUint(raw, 10, 64)
		if err == nil {
			categoryID = &cid
		}
	}
	var branchID *uint64
	if raw := c.Query("branch_id"); raw != "" {
		bid, err := strconv.ParseUint(raw, 10, 64)
		if err == nil {
			branchID = &bid
		}
	}
	activeOnly := c.Query("active") == "1" || c.Query("active") == "true"

	items, err := h.repo.ListItems(c.Request.Context(), orgID, categoryID, branchID, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// GetItem handles GET /api/v1/menu/items/:id.
func (h *Handler) GetItem(c *gin.Context) {
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
	item, err := h.repo.FindItemByID(c.Request.Context(), id, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "item tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateItem handles POST /api/v1/menu/items.
func (h *Handler) CreateItem(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	item, err := h.repo.CreateItem(c.Request.Context(), orgID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// UpdateItem handles PATCH /api/v1/menu/items/:id.
func (h *Handler) UpdateItem(c *gin.Context) {
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
	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	item, err := h.repo.UpdateItem(c.Request.Context(), id, orgID, &req)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "item tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// UpsertBranchSetting handles PUT /api/v1/menu/items/:id/branches/:branch_id.
func (h *Handler) UpsertBranchSetting(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item id tidak valid"})
		return
	}
	branchID, err := strconv.ParseUint(c.Param("branch_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch id tidak valid"})
		return
	}
	var req UpsertBranchSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	bs, err := h.repo.UpsertBranchSetting(c.Request.Context(), itemID, orgID, branchID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bs)
}

// GetBranchSettings handles GET /api/v1/menu/items/:id/branches.
func (h *Handler) GetBranchSettings(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item id tidak valid"})
		return
	}
	settings, err := h.repo.GetBranchSettings(c.Request.Context(), itemID, orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": settings})
}

// RegisterRoutes mounts menu endpoints.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	menu := v1.Group("/menu", authMiddleware)
	{
		menu.GET("/categories", h.ListCategories)
		menu.POST("/categories", h.CreateCategory)
		menu.PATCH("/categories/:id", h.UpdateCategory)

		menu.GET("/items", h.ListItems)
		menu.POST("/items", h.CreateItem)
		menu.GET("/items/:id", h.GetItem)
		menu.PATCH("/items/:id", h.UpdateItem)
		menu.GET("/items/:id/branches", h.GetBranchSettings)
		menu.PUT("/items/:id/branches/:branch_id", h.UpsertBranchSetting)
	}
}
