package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	r   *redis.Client
	ctx context.Context
}

func NewCache(connString string) (Cache, error) {
	r := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	err := r.Ping(context.Background()).Err()
	if err != nil {
		return Cache{}, err
	}
	return Cache{
		r:   r,
		ctx: context.Background(),
	}, nil
}
