package redis

import "github.com/redis/go-redis/v9"

func Newredis(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}