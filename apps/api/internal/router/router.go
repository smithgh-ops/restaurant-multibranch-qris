package router

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/auth"
	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/branch"
	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/config"
	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/handler"
	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/menu"
	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/organization"
)

// New creates and returns the root Gin engine with all routes registered.
func New(cfg *config.Config, db *sql.DB) *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.CORSOrigins))

	// Health check (unauthenticated)
	r.GET("/health", handler.Health)

	// Auth middleware factory
	authMW := middleware.Authenticate(cfg.JWTSecret)

	// Auth service
	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, auth.Config{
		JWTSecret:       cfg.JWTSecret,
		AccessTokenTTL:  time.Duration(cfg.AccessTokenMinutes) * time.Minute,
		RefreshTokenTTL: time.Duration(cfg.RefreshTokenDays) * 24 * time.Hour,
	})
	authHandler := auth.NewHandler(authSvc)

	// Organization handler
	orgRepo := organization.NewRepository(db)
	orgHandler := organization.NewHandler(orgRepo)

	// Branch handler
	branchRepo := branch.NewRepository(db)
	branchHandler := branch.NewHandler(branchRepo)

	// Menu handler
	menuRepo := menu.NewRepository(db)
	menuHandler := menu.NewHandler(menuRepo)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		v1.GET("/info", handler.Info(cfg))

		auth.RegisterRoutes(v1, authHandler, authMW)
		organization.RegisterRoutes(v1, orgHandler, authMW)
		branch.RegisterRoutes(v1, branchHandler, authMW)
		menu.RegisterRoutes(v1, menuHandler, authMW)
	}

	return r
}
