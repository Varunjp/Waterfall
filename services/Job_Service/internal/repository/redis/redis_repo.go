package redisRepo

import (
	"context"
	"fmt"
	"job_service/internal/pkg/grpc/interceptor"
	"job_service/internal/repository"
	"time"

	"github.com/go-redis/redis"
)

type RedisRepo struct {
	redis     *redis.Client
	adminRepo repository.AdminRepository
}

func NewRedisRepo(r *redis.Client, a repository.AdminRepository) *RedisRepo {
	return &RedisRepo{redis: r, adminRepo: a}
}

func (r *RedisRepo) CheckQuota(ctx context.Context, appID string) error {
	freeLimit, freeUsage, err := r.getFreeQuota(ctx, appID)
	if err != nil {
		return err
	}

	if freeUsage < freeLimit {
		return nil
	}

	limit, err := r.getPlanLimit(ctx, appID)
	if err != nil {
		return err
	}

	usage, err := r.getMonthlyUsage(ctx, appID)
	if err != nil {
		return err
	}

	if usage >= limit {
		return interceptor.ErrQuotaExceeded
	}

	return nil
}

func (r *RedisRepo) Incr(ctx context.Context, appID string) error {

	freeLimit, freeUsage, err := r.getFreeQuota(ctx, appID)
	freeUsageKey := fmt.Sprintf("free-usage:%s", appID)

	if err != nil {
		return err
	}

	if freeUsage < freeLimit {
		err = r.redis.Incr(freeUsageKey).Err()

		if err != nil {
			return err
		}

		return r.redis.Expire(freeUsageKey, 24*time.Hour).Err()
	}

	// Subcribed plan increment

	limit, err := r.getPlanLimit(ctx, appID)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("usage:%s:%s", appID, time.Now().Format("2006-01"))

	usage, err := r.redis.Incr(key).Result()
	if err != nil {
		return err
	}

	err = r.redis.Expire(key, 24*time.Hour).Err()
	if err != nil {
		return err
	}

	if int(usage) <= limit {
		return nil
	}

	return nil
}

func (r *RedisRepo) getPlanLimit(ctx context.Context, appID string) (int, error) {
	planKey := fmt.Sprintf("plan:%s", appID)
	limit, err := r.redis.Get(planKey).Int()
	if err == redis.Nil {
		planID, err := r.adminRepo.GetPlanID(ctx, appID)
		if err != nil {
			return 0, err
		}
		name, limit, err := r.adminRepo.GetPlanDetails(ctx, planID)
		if err != nil {
			return 0, err
		}
		if name == "FREE" {
			return 0, interceptor.ErrQuotaExceeded
		}
		err = r.redis.Set(planKey, any(limit), 24*time.Hour).Err()
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	return limit, nil
}

func (r *RedisRepo) getMonthlyUsage(ctx context.Context, appID string) (int, error) {
	usageKey := fmt.Sprintf("usage:%s:%s", appID, time.Now().Format("2006-01"))
	usage, err := r.redis.Get(usageKey).Int()
	if err == redis.Nil {
		usage, err = r.adminRepo.GetMonthlyUsage(ctx, appID)
		if err != nil {
			return 0, err
		}
		err = r.redis.Set(usageKey, any(usage), 1*24*time.Hour).Err()
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	return usage, nil
}

func (r *RedisRepo) getFreeQuota(ctx context.Context, appID string) (int, int, error) {
	freeLimitKey := fmt.Sprintf("free-limit:%s", appID)
	freeUsageKey := fmt.Sprintf("free-usage:%s", appID)

	freeLimit, limitErr := r.redis.Get(freeLimitKey).Int()
	freeUsage, usageErr := r.redis.Get(freeUsageKey).Int()

	if limitErr == nil && usageErr == nil {
		return freeLimit, freeUsage, nil
	}

	if limitErr != nil && limitErr != redis.Nil {
		return 0, 0, limitErr
	}

	if usageErr != nil && usageErr != redis.Nil {
		return 0, 0, usageErr
	}

	freeLimit, freeUsage, err := r.adminRepo.GetFreeQuota(ctx, appID)
	if err != nil {
		return 0, 0, err
	}

	err = r.redis.SetNX(freeLimitKey, any(freeLimit), 24*time.Hour).Err()
	if err != nil {
		return 0, 0, err
	}

	err = r.redis.SetNX(freeUsageKey, any(freeUsage), 24*time.Hour).Err()
	if err != nil {
		return 0, 0, err
	}

	freeLimit, err = r.redis.Get(freeLimitKey).Int()
	if err != nil {
		return 0, 0, err
	}

	freeUsage, err = r.redis.Get(freeUsageKey).Int()
	if err != nil {
		return 0, 0, err
	}

	return freeLimit, freeUsage, nil
}
