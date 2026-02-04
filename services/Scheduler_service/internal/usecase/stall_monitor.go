package usecase

import (
	"context"
	"encoding/json"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/producer"
	redisClient "scheduler_service/internal/redis"
	"time"
)

type StallMonitor struct {
	redis *redisClient.Client
	producer *producer.KafkaProducer
}

func NewStallMonitor(r *redisClient.Client, p *producer.KafkaProducer) *StallMonitor {
	return &StallMonitor{r,p}
}

func (s *StallMonitor) Run(ctx context.Context) {
	ticker := time.NewTicker(10 *time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return 
		case <-ticker.C:
			s.scan(ctx)
		}
	}
}

func (s *StallMonitor) scan(ctx context.Context) {
	keys,_ := s.redis.Keys(ctx,"heartbeat:*").Result()

	now := time.Now().Unix()

	for _,hbKey := range keys {
		jobID := hbKey[len("heartbeat:"):]

		ts,err := s.redis.Get(ctx,hbKey).Int64()
		if err != nil {
			continue 
		}

		if now-ts > 30 {
			s.handlestall(ctx,jobID)
		}
	}
}

func (s *StallMonitor) handlestall(ctx context.Context,jobID string) {
	metaKey := "running:"+jobID 

	raw,err := s.redis.Get(ctx,metaKey).Result()

	if err != nil {
		return 
	}

	var job domain.Job
	json.Unmarshal([]byte(raw),&job)

	job.Retry++

	if job.Retry > job.MaxRetries {

		s.producer.Publish(ctx,map[string]any{
			"job_id":job.JobID,
			"app_id":job.AppID,
			"Status": "DLQ",
			"retries": job.Retry,
		})

		return 
	}

	s.producer.Publish(ctx,map[string]any{
		"job_id": job.JobID,
		"status": "RETRYING",
	})

	s.redis.LPush(ctx,"job:ingress",raw)
}