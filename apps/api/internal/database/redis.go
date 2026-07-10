package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/config"
)

// NewRedis creates a Redis client and verifies connectivity.
func NewRedis(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis: ping: %w", err)
	}

	slog.Info("Redis connected", "addr", cfg.RedisAddr)
	return rdb, nil
}
