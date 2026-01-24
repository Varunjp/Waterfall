package repository

import (
	"context"
	"encoding/json"
	"errors"
	"scheduler_service/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	workerTTL = 20 * time.Second
)

type RedisWorkerRegistry struct {
	rdb *redis.Client
}

func NewRedisWorkerRegistry(rdb *redis.Client) *RedisWorkerRegistry {
	return &RedisWorkerRegistry{rdb: rdb}
}

func (r *RedisWorkerRegistry) workerKey(id string) string {
	return "workers:"+id 
}

func (r *RedisWorkerRegistry) Register(worker domain.Worker) error {
	worker.ActiveJobs = 0
	worker.LastSeen = time.Now()

	data,_ := json.Marshal(worker)
	return r.rdb.Set(
		context.Background(),
		r.workerKey(worker.WorkerID),
		data, 
		workerTTL,
	).Err()
}

func (r *RedisWorkerRegistry) Heartbeat(workerID string) error {
	return r.rdb.Expire(
		context.Background(),
		r.workerKey(workerID),
		workerTTL,
	).Err()
}

func (r *RedisWorkerRegistry) Acquire(appID, jobType string) (*domain.Worker,error) {
	ctx := context.Background()

	keys,err := r.rdb.Keys(ctx,"workers:*").Result()
	if err != nil {
		return nil,err 
	}

	for _,key := range keys {
		raw,err := r.rdb.Get(ctx,key).Result()
		if err != nil {
			continue
		}

		var w domain.Worker
		if err := json.Unmarshal([]byte(raw),&w); err != nil {
			continue 
		}

		if w.AppID != appID {
			continue
		}
		if w.ActiveJobs >= w.Concurrency {
			continue
		}
		if !hasCapability(w.Capabilities,jobType) {
			continue
		}

		w.ActiveJobs++
		w.LastSeen = time.Now()

		updated,_ := json.Marshal(w)
		err = r.rdb.Set(ctx, key, updated, workerTTL).Err()
		if err != nil {
			return nil,err 
		}

		return &w,nil 
	}

	return nil, errors.New("no eligible worker available")
}

func (r *RedisWorkerRegistry) Release(workerID string) error {
	ctx := context.Background()
	key := r.workerKey(workerID)

	raw,err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return err 
	}

	var w domain.Worker
	if err := json.Unmarshal([]byte(raw),&w); err != nil {
		return err 
	}

	if w.ActiveJobs > 0 {
		w.ActiveJobs--
	}

	data,_ := json.Marshal(w)
	return r.rdb.Set(ctx,key,data,workerTTL).Err()
}

func hasCapability(caps []string, jobType string) bool {
	for _,c := range caps {
		if c == jobType {
			return true 
		}
	}
	return false 
}