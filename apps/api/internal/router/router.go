package router

import (
	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/config"
	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/handler"
	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
)

// New creates and returns the root Gin engine with all routes registered.
func New(cfg *config.Config) *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.CORSOrigins))

	// Health check (unauthenticated)
	r.GET("/health", handler.Health)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		v1.GET("/info", handler.Info(cfg))
		// Future route groups will be registered here:
		// auth.RegisterRoutes(v1, deps)
		// branches.RegisterRoutes(v1, deps)
		// menus.RegisterRoutes(v1, deps)
		// orders.RegisterRoutes(v1, deps)
		// payments.RegisterRoutes(v1, deps)
		// kitchen.RegisterRoutes(v1, deps)
		// reports.RegisterRoutes(v1, deps)
	}

	return r
}
