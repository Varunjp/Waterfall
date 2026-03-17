package redisRepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"job_service/internal/repository"
	"time"

	"github.com/go-redis/redis"
)

type RedisRepo struct {
	redis *redis.Client
	adminRepo repository.AdminRepository
}

func NewRedisRepo(r *redis.Client,a repository.AdminRepository) *RedisRepo {
	return &RedisRepo{redis: r,adminRepo: a}
}

func (r *RedisRepo) CheckQuota(ctx context.Context,appID string) error {

	planKey := fmt.Sprintf("plan:%s",appID)
	limit,err := r.redis.Get(planKey).Int()
	if err == redis.Nil {
		planID,err := r.adminRepo.GetPlanID(ctx,appID)
		if err != nil {
			return err 
		}
		limit,err = r.adminRepo.GetPlanDetails(ctx,planID)
		if err != nil {
			return err 
		}
		err = r.redis.Set(planKey,any(limit),24*time.Hour).Err()
		if err != nil {
			return err 
		}
	}else if err != nil {
		return err 
	}

	usageKey := fmt.Sprintf("usage:%s:%s",appID,time.Now().Format("2006-01"))
	usage,err := r.redis.Get(usageKey).Int()
	if err == redis.Nil {
		usage,err = r.adminRepo.GetMonthlyUsage(ctx,appID)
		if errors.Is(err,sql.ErrNoRows) {
			usage = 0
		}else if err != nil {
			return err
		}
		err = r.redis.Set(usageKey,any(usage),1*24*time.Hour).Err()
		if err != nil {
			return err 
		}
	}

	if usage >= limit {
		return errors.New("quota exceeded")
	}

	return nil 
}

func (r *RedisRepo) Incr(ctx context.Context,appID string)error {

	key := fmt.Sprintf("usage:%s:%s",appID,time.Now().Format("2006-01"))

	err := r.redis.Incr(key).Err()

	return err 
}