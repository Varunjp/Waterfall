package client

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedis(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func Ping(rdb *redis.Client) error {
	return rdb.Ping(context.Background()).Err()
}