package usecase

import (
	"context"
	"errors"
	"fmt"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/metrics"
	"scheduler_service/internal/monitoring"
	"scheduler_service/internal/producer"
	redisClient "scheduler_service/internal/redis"
	"scheduler_service/internal/repository"
	"time"
)

var (
	ErrNoJobAvailable = errors.New("no job available")
	ErrQuotaExceeded  = errors.New("app quota exceeded")
)

type Assigner struct {
	redis     *redisClient.Client
	adminRepo *repository.AdminRepo
	metrics   *metrics.SchedulerMetrics
	producer  *producer.KafkaProducer
	runtime   *monitoring.Store
}

func NewAssigner(
	r *redisClient.Client,
	a *repository.AdminRepo,
	m *metrics.SchedulerMetrics,
	p *producer.KafkaProducer,
	store *monitoring.Store,
) *Assigner {
	return &Assigner{
		redis:     r,
		adminRepo: a,
		metrics:   m,
		producer:  p,
		runtime:   store,
	}
}

func (a *Assigner) Assign(ctx context.Context, job domain.Job) error {
	limit, err := a.adminRepo.Concurrency(ctx, job.AppID)
	if err != nil {
		return err
	}

	if limit == 0 {
		return ErrQuotaExceeded
	}
	stream := fmt.Sprintf("stream:jobs:%s:%s", job.AppID, job.Type)
	group := fmt.Sprintf("workers:%s:%s", job.AppID, job.Type)

	if err := redisClient.EnsureGroup(ctx, a.redis.Client, stream, group); err != nil {
		return err
	}

	key := fmt.Sprintf("concurrency:%s", job.AppID)

	_, err = a.redis.EvalSha(
		ctx,
		a.redis.AssignSHA,
		[]string{key, stream},
		limit,
		job.JobID,
		job.Payload,
		job.Retry,
		job.ManualRetry,
	).Result()

	if err != nil {
		a.metrics.JobsFailed.Inc()
		return err
	}

	a.metrics.RunningJobs.Inc()
	a.metrics.JobsAssigned.Inc()
	if a.runtime != nil {
		_ = a.runtime.RecordQueuedJob(ctx, job, time.Now().UTC())
	}
	return nil
}
