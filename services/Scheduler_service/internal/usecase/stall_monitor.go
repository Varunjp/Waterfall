package usecase

import (
	"context"
	"encoding/json"
	"log"
	"scheduler_service/internal/consumer"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/monitoring"
	"scheduler_service/internal/producer"
	redisClient "scheduler_service/internal/redis"
	"time"

	"github.com/redis/go-redis/v9"
)

type StallMonitor struct {
	redis          *redisClient.Client
	runnerConsumer *consumer.JobCreatedConsumer
	producer       *producer.KafkaProducer
	runtime        *monitoring.Store
}

func NewStallMonitor(
	r *redisClient.Client,
	rc *consumer.JobCreatedConsumer,
	p *producer.KafkaProducer,
	store *monitoring.Store,
) *StallMonitor {
	return &StallMonitor{
		redis:          r,
		runnerConsumer: rc,
		producer:       p,
		runtime:        store,
	}
}

func (s *StallMonitor) Run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msg, err := s.runnerConsumer.Read(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Println("kafka read failed", err)
				continue
			}

			s.scan(ctx, msg)
		}
	}
}

func (s *StallMonitor) scan(ctx context.Context, raw []byte) {

	var job domain.Job
	if err := json.Unmarshal(raw, &job); err != nil {
		log.Println("invalid job payload", err)
		return
	}

	hbKey := "heartbeat:" + job.JobID
	ts, err := s.redis.Get(ctx, hbKey).Int64()

	if err == redis.Nil {
		s.handlestall(ctx, job.JobID, job)
		return
	}

	if err != nil {
		return
	}

	now := time.Now().Unix()

	if now-ts > 30 {
		s.handlestall(ctx, job.JobID, job)
	}
}

func (s *StallMonitor) handlestall(ctx context.Context, jobID string, stjob domain.Job) {
	job, err := s.stalledJob(ctx, jobID, stjob)
	if err != nil {
		return
	}

	if s.runtime != nil {
		_ = s.runtime.ReleaseJob(ctx, jobID, time.Now().UTC())
	}

	if job.AppID != "" {
		_ = s.redis.Decr(ctx, "concurrency:"+job.AppID)
	}

	_, _ = s.redis.Del(ctx, "heartbeat:"+jobID, "running:"+jobID).Result()

	job.Retry++
	s.publishStallOutcome(ctx, job)
}

func (s *StallMonitor) stalledJob(ctx context.Context, jobID string, fallback domain.Job) (domain.Job, error) {
	raw, err := s.redis.Get(ctx, "running:"+jobID).Result()
	if err == redis.Nil {
		return fallback, nil
	}

	if err != nil {
		return domain.Job{}, err
	}

	var job domain.Job
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		return domain.Job{}, err
	}

	if job.JobID == "" {
		job.JobID = fallback.JobID
	}
	if job.AppID == "" {
		job.AppID = fallback.AppID
	}
	if job.MaxRetries == 0 && fallback.MaxRetries > 0 {
		job.MaxRetries = fallback.MaxRetries
	}

	return job, nil
}

func (s *StallMonitor) publishStallOutcome(ctx context.Context, job domain.Job) {
	event := map[string]any{
		"job_id": job.JobID,
		"app_id": job.AppID,
		"retry":  job.Retry,
	}

	if job.Retry > job.MaxRetries {
		event["status"] = "DLQ"
		_ = s.producer.Publish(ctx, event)
		return
	}

	event["status"] = domain.JobRetry
	_ = s.producer.Publish(ctx, event)
}
