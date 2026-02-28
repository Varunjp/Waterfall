package usecase

import (
	"context"
	"errors"
	"fmt"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/metrics"
	"scheduler_service/internal/producer"
	redisClient "scheduler_service/internal/redis"
	"scheduler_service/internal/repository"
)

var (
	ErrNoJobAvailable = errors.New("no job available")
	ErrQuotaExceeded  = errors.New("app quota exceeded")
)

type Assigner struct {
	redis *redisClient.Client
	adminRepo *repository.AdminRepo
	metrics *metrics.SchedulerMetrics
	producer *producer.KafkaProducer
}

func NewAssigner(r *redisClient.Client,a *repository.AdminRepo, m *metrics.SchedulerMetrics, p  *producer.KafkaProducer) *Assigner {
	return &Assigner{redis: r,adminRepo: a,metrics: m, producer: p}
}

func (a *Assigner) Assign(ctx context.Context,job domain.Job) error {
	limit,err := a.adminRepo.Concurrency(job.AppID)
	if err != nil {
		return err 
	}

	if limit == 0 {
		return ErrQuotaExceeded
	}
	stream := fmt.Sprintf("stream:jobs:%s:%s",job.AppID,job.Type)
	group  := fmt.Sprintf("workers:%s:%s", job.AppID, job.Type)

	if err := redisClient.EnsureGroup(ctx,a.redis.Client,stream,group); err!= nil  {
		return err 
	}

	key := fmt.Sprintf("concurrency:%s",job.AppID)

	_,err = a.redis.EvalSha(
		ctx,
		a.redis.AssignSHA,
		[]string{key,stream},
		limit,
		job.JobID,
		job.Payload,
		job.Retry,
		job.ManualRetry,
	).Result()
	
	a.metrics.RunningJobs.Inc()

	if err != nil {
		a.metrics.JobsFailed.Inc()
		return err 
	}

	// a.producer.Publish(ctx,map[string]any{
	// 	"job_id":job.JobID,
	// 	"status": "RUNNING",
	// })

	a.metrics.JobsAssigned.Inc()
	return nil 
}