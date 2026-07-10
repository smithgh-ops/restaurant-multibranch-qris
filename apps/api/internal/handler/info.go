package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/config"
)

// InfoResponse is the shape returned by GET /api/v1/info.
type InfoResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Env     string `json:"env"`
}

// Info handles GET /api/v1/info and returns basic application metadata.
func Info(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, InfoResponse{
			Name:    cfg.AppName,
			Version: "0.1.0",
			Env:     cfg.AppEnv,
		})
	}
}
