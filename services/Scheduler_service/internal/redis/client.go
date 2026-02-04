package redisClient

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
	AssignSHA string 
}

func NewRedisClient(addr string) *Client {
	r := redis.NewClient(&redis.Options{Addr: addr})

	lua,_ := os.ReadFile("internal/redis/lua/assign_job.lua")
	sha,_ := r.ScriptLoad(context.Background(),string(lua)).Result()

	return &Client{Client: r,AssignSHA: sha}
}