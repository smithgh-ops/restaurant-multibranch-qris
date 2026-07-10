package auth

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts the auth endpoints on the given router group.
// The authMiddleware parameter is the JWT auth middleware defined in middleware/auth.go.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	auth := v1.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshHandler)
		auth.POST("/logout", h.Logout)
		auth.GET("/me", authMiddleware, h.Me)
	}
}
