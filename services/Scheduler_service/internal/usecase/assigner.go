package usecase

import (
	"context"
	"fmt"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/metrics"
	"scheduler_service/internal/monitoring"
	"scheduler_service/internal/producer"
	redisClient "scheduler_service/internal/redis"
	"time"
)

type Assigner struct {
	redis    *redisClient.Client
	metrics  *metrics.SchedulerMetrics
	producer *producer.KafkaProducer
	runtime  *monitoring.Store
}

func NewAssigner(
	r *redisClient.Client,
	m *metrics.SchedulerMetrics,
	p *producer.KafkaProducer,
	store *monitoring.Store,
) *Assigner {
	return &Assigner{
		redis:    r,
		metrics:  m,
		producer: p,
		runtime:  store,
	}
}

func (a *Assigner) Assign(ctx context.Context, job domain.Job) error {
	stream := fmt.Sprintf("stream:jobs:%s:%s", job.AppID, job.Type)
	group := fmt.Sprintf("workers:%s:%s", job.AppID, job.Type)

	if err := redisClient.EnsureGroup(ctx, a.redis.Client, stream, group); err != nil {
		return err
	}

	key := fmt.Sprintf("concurrency:%s", job.AppID)

	_, err := a.redis.EvalSha(
		ctx,
		a.redis.AssignSHA,
		[]string{key, stream},
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
