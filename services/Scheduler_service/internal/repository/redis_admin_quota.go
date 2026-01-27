package repository

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const (
	defaultLimit = 5000
)

type RedisAdminQuotaRepo struct {
	rdb *redis.Client
}

func NewRedisAdminQuotaRepo(rdb *redis.Client) *RedisAdminQuotaRepo {
	return &RedisAdminQuotaRepo{rdb: rdb}
}

func (r *RedisAdminQuotaRepo) key(appID string) string {
	return "quota:"+appID
}

func (r *RedisAdminQuotaRepo) CanStart(appID string)(bool,error) {
	ctx := context.Background()

	val,err := r.rdb.Get(ctx,r.key(appID)).Result()
	if err == redis.Nil {
		// need to add logic for getting details from admin db
		return true,nil 
	}

	if err != nil {
		return false,err 
	}

	running,_ := strconv.Atoi(val)
	// correct logic for getting limit
	return running < defaultLimit,nil 
}

func (r *RedisAdminQuotaRepo) Increment(appID string)error{
	ctx := context.Background()
	return r.rdb.Incr(ctx,r.key(appID)).Err()
}

func (r *RedisAdminQuotaRepo) Decrement(appID string) error {
	ctx := context.Background()
	return r.rdb.Decr(ctx,r.key(appID)).Err()
}