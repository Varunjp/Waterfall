package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/repository/lua"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisJobStore struct {
	rdb *redis.Client
}

func NewRedisJobStore(rdb *redis.Client) *RedisJobStore {
	return &RedisJobStore{rdb: rdb}
}

func (r *RedisJobStore) pendingKey(appID,jobType string) string {
	return fmt.Sprintf("pending:%s:%s",appID,jobType)
}

func (r *RedisJobStore) seenKey(jobID string) string {
	return "job:seen:"+jobID
}

func (r *RedisJobStore) SavePendingJob(job domain.Job) error {
	ctx := context.Background()

	ok,err := r.rdb.SetNX(ctx, r.seenKey(job.JobID),1, 24 * time.Hour).Result()
	if err != nil || !ok {
		return err 
	}

	payload,err := json.Marshal(job)
	if err != nil {
		return err 
	}
	return r.rdb.RPush(
		ctx,
		r.pendingKey(job.AppID,job.Type),
		payload,
	).Err()
}

func (r *RedisJobStore) PollJob(
	appID string,
	jobType string,
	workerID string,
)(*domain.Job,error) {

	ctx := context.Background()

	pendingKey := fmt.Sprintf("pending:%s:%s",appID,jobType)
	runningKeyPrefix := "running:"
	now := time.Now().Unix()

	res, err := r.rdb.Eval(
		ctx,
		lua.PollJobLua,
		[]string{pendingKey,runningKeyPrefix},
		workerID,
		60,
		now,
	).Result()
	
	if err != nil || res == nil {
		return nil,err 
	}

	var job domain.Job
	err = json.Unmarshal([]byte(res.(string)),&job)
	if err != nil {
		return nil,err 
	}
	return &job,nil 
}

func (r *RedisJobStore) RefershRunning(jobID string) error {
	ctx := context.Background()

	return r.rdb.Expire(
		ctx,
		"running:"+jobID,
		60*time.Second,
	).Err()
}

func (r *RedisJobStore) DeleteRunning(jobID string) error {
	ctx := context.Background()
	return r.rdb.Del(ctx, "running:"+jobID).Err()
}

func (r *RedisJobStore) Requeue(job *domain.Job) error {
	job.Retry++
	return r.SavePendingJob(*job)
}

func (r *RedisJobStore) FindStalled(timeout time.Duration)([]domain.RunningJob,error) {
	ctx := context.Background()
	keys, err := r.rdb.Keys(ctx, "running:*").Result()
	if err != nil {
		return nil,err 
	}
	
	var stalled []domain.RunningJob

	for _,key := range keys {
		ttl,err := r.rdb.TTL(ctx,key).Result()
		if err != nil || ttl > 0 {
			continue 
		}

		data,err := r.rdb.HGetAll(ctx,key).Result()
		if err != nil || len(data) == 0 {
			continue 
		}

		started,_ := strconv.ParseInt(data["started_at"],10,64)
		lastBeat, _ := strconv.ParseInt(data["last_beat"],10,64)

		stalled = append(stalled, domain.RunningJob{
			JobID: data["job_id"],
			AppID: data["app_id"],
			JobType: data["job_type"],
			WorkerID: data["worker_id"],
			StartedAt: time.Unix(started,0),
			LastBeat: time.Unix(lastBeat,0),
		})
	}

	return stalled,nil 
}