package redis

import (
	"context"
	"time"

	"github.com/mormm/boxing/internal/platform/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	*redis.Client
}

func New(cfg *config.Config) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Redis{Client: rdb}, nil
}

func (r *Redis) Close() error {
	return r.Client.Close()
}