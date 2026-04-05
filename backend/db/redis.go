package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client

func InitRedis() error {
	RDB = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	if err := RDB.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	// structured log — shows which address Redis connected to
	slog.Info("redis connected", "addr", os.Getenv("REDIS_ADDR"))

	return nil
}