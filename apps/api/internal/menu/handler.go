package menu

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
)

// Handler groups HTTP handlers for menu endpoints.
type Handler struct {
	repo          *Repository
	uploadDir     string
	publicBaseURL string
}

// NewHandler creates a new menu handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// WithUpload configures the handler with upload directory and public base URL.
func (h *Handler) WithUpload(uploadDir, publicBaseURL string) *Handler {
	h.uploadDir = uploadDir
	h.publicBaseURL = publicBaseURL
	return h
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

	categoryID, err := parseOptionalUint64Query(c.Query("category_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameter category_id tidak valid"})
		return
	}
	branchID, err := parseOptionalUint64Query(c.Query("branch_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameter branch_id tidak valid"})
		return
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

// UploadItemImage handles POST /api/v1/menu/items/:id/image.
// It accepts a multipart/form-data file with field name "image".
// Allowed types: jpeg, jpg, png, webp. Max size: 5 MB.
func (h *Handler) UploadItemImage(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	// Limit request size to 5 MB.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<20)
	if err := c.Request.ParseMultipartForm(5 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ukuran file melebihi batas 5 MB"})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "field 'image' tidak ditemukan"})
		return
	}
	defer file.Close()

	// Validate MIME type by reading the first 512 bytes.
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mimeType := http.DetectContentType(buf[:n])
	allowedMIME := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}
	ext, allowed := allowedMIME[mimeType]
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tipe file tidak didukung; gunakan jpg, png, atau webp"})
		return
	}

	// Ensure upload directory exists.
	if err := os.MkdirAll(h.uploadDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Build a safe filename: menu_<itemID>_<originalbase><ext>
	origBase := strings.TrimSuffix(filepath.Base(header.Filename), filepath.Ext(header.Filename))
	// Strip any non-alphanumeric characters from the original name.
	safe := ""
	for _, r := range origBase {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			safe += string(r)
		}
	}
	if len(safe) > 40 {
		safe = safe[:40]
	}
	if safe == "" {
		safe = "img"
	}
	filename := fmt.Sprintf("menu_%d_%s%s", itemID, safe, ext)
	destPath := filepath.Join(h.uploadDir, filename)

	// Delete any previous file for this item to avoid orphans.
	if oldURL, _ := h.repo.ClearItemImage(c.Request.Context(), itemID, orgID); oldURL != "" {
		// Extract just the filename portion from the stored URL.
		oldFile := filepath.Join(h.uploadDir, filepath.Base(oldURL))
		_ = os.Remove(oldFile)
	}

	// Write new file.
	dst, err := os.Create(destPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer dst.Close()

	// Write the already-read bytes first, then the remainder.
	if _, err := dst.Write(buf[:n]); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Build public URL and persist to DB.
	publicURL := strings.TrimRight(h.publicBaseURL, "/") + "/uploads/" + filename
	item, err := h.repo.SetItemImage(c.Request.Context(), itemID, orgID, publicURL)
	if errors.Is(err, sql.ErrNoRows) {
		_ = os.Remove(destPath)
		c.JSON(http.StatusNotFound, gin.H{"error": "item tidak ditemukan"})
		return
	}
	if err != nil {
		_ = os.Remove(destPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// DeleteItemImage handles DELETE /api/v1/menu/items/:id/image.
func (h *Handler) DeleteItemImage(c *gin.Context) {
	orgID, ok := middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}
	oldURL, err := h.repo.ClearItemImage(c.Request.Context(), itemID, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "item tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if oldURL != "" {
		oldFile := filepath.Join(h.uploadDir, filepath.Base(oldURL))
		_ = os.Remove(oldFile)
	}
	c.JSON(http.StatusOK, gin.H{"message": "gambar berhasil dihapus"})
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
		menu.POST("/items/:id/image", h.UploadItemImage)
		menu.DELETE("/items/:id/image", h.DeleteItemImage)
		menu.GET("/items/:id/branches", h.GetBranchSettings)
		menu.PUT("/items/:id/branches/:branch_id", h.UpsertBranchSetting)
	}
}

func parseOptionalUint64Query(raw string) (*uint64, error) {
	if raw == "" {
		return nil, nil
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
