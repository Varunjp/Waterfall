package consumer

import (
	"context"
	"encoding/json"
	"log"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/repository"

	"github.com/segmentio/kafka-go"
)

type JobCreatedConsumer struct {
	reader *kafka.Reader
	jobStore   *repository.RedisJobStore
}

func NewJobCreatedConsumer(reader *kafka.Reader, jobStore *repository.RedisJobStore) *JobCreatedConsumer {
	return &JobCreatedConsumer{
		reader: reader,
		jobStore: jobStore,
	}
}

func (c *JobCreatedConsumer) Start(ctx context.Context) error {
	defer c.reader.Close()
	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler consumer stopped")
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				log.Println("kafka read error:",err)
				continue
			}
			
			var evt JobCreatedEvent 
			if err := json.Unmarshal(msg.Value,&evt); err != nil {
				continue 
			}
			
			job := domain.Job {
				JobID: evt.JobID,
				AppID: evt.AppID,
				Type: evt.Type,
				Payload: string(evt.Payload),
				Retry: evt.Retry,
				MaxRetries: evt.MaxRetries,
			}

			if err := c.jobStore.SavePendingJob(job); err != nil {
				log.Println("redis save error:",err)
			}
		}
	}
}