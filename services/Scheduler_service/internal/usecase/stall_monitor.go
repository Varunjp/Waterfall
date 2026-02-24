package usecase

import (
	"context"
	"encoding/json"
	"log"
	"scheduler_service/internal/consumer"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/producer"
	redisClient "scheduler_service/internal/redis"
	"time"

	"github.com/redis/go-redis/v9"
)

type StallMonitor struct {
	redis *redisClient.Client
	runnerConsumer *consumer.JobCreatedConsumer
	producer *producer.KafkaProducer
}

func NewStallMonitor(r *redisClient.Client, rc *consumer.JobCreatedConsumer ,p *producer.KafkaProducer) *StallMonitor {
	return &StallMonitor{redis: r,runnerConsumer: rc,producer: p}
}

func (s *StallMonitor) Run(ctx context.Context) {
	ticker := time.NewTicker(10 *time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return 
		case <-ticker.C:
			msg,err := s.runnerConsumer.Read(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return 
				}
				log.Println("kafka read failed",err)
				continue
			}

			s.scan(ctx,msg)
		}
	}
}

func (s *StallMonitor) scan(ctx context.Context,raw []byte) {

	var job domain.Job
	if err := json.Unmarshal(raw,&job); err != nil {
		log.Println("invalid job payload",err)
		return 
	}

	hbKey := "heartbeat:"+job.JobID
	ts,err := s.redis.Get(ctx,hbKey).Int64()

	if err == redis.Nil {
		s.handlestall(ctx,job.JobID)
		return 
	}

	if err != nil {
		return 
	}

	now := time.Now().Unix()

	if now-ts > 30 {
		s.handlestall(ctx,job.JobID)
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
		"status": "JOB_RETRY",
		"retry": job.Retry,
	})

	s.redis.LPush(ctx,"job:ingress",raw)
}