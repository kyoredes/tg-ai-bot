package storage

import (
	"context"
	"gateway/internal/config"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(cfg *config.RedisConfig) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		DB:       cfg.DB,
		Password: cfg.Password,
	})
	return &RedisClient{Client: client}
}
