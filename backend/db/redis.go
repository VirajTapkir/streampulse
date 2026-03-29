package db

import (
    "context"
    "fmt"
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

    return nil
}