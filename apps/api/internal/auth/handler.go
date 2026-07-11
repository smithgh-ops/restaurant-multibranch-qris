package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler groups HTTP handlers for the auth endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new auth handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email dan password wajib diisi"})
		return
	}

	pair, _, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email atau password salah"})
		return
	}
	if errors.Is(err, ErrAccountInactive) {
		c.JSON(http.StatusForbidden, gin.H{"error": "akun tidak aktif"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, pair)
}

// RefreshHandler handles POST /api/v1/auth/refresh.
func (h *Handler) RefreshHandler(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token wajib diisi"})
		return
	}

	pair, err := h.svc.Refresh(c.Request.Context(), body.RefreshToken)
	if errors.Is(err, ErrTokenInvalid) || errors.Is(err, ErrAccountInactive) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sesi tidak valid, silakan login kembali"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, pair)
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handler) Logout(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&body)

	if err := h.svc.Logout(c.Request.Context(), body.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "berhasil logout"})
}

// Me handles GET /api/v1/auth/me (requires authentication).
func (h *Handler) Me(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	me, err := h.svc.Me(c.Request.Context(), userID.(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, me)
}

// ChangePassword handles PATCH /api/v1/auth/me/password (requires authentication).
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID.(uint64), req.CurrentPassword, req.NewPassword); errors.Is(err, ErrWrongPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "password saat ini tidak sesuai"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password berhasil diubah"})
}
