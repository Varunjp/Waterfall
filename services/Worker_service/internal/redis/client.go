package redisClient

import "github.com/go-redis/redis"

func NewRedis(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}