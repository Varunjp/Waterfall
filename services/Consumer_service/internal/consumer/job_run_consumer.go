package consumer

import (
	"consumer_service/internal/domain"
	"consumer_service/internal/usecase"
	"context"
	"encoding/json"
	"errors"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type JobRunConsumer struct {
	reader *kafka.Reader
	usecase *usecase.UpdateJobStatusUsecase
	logger  *zap.Logger
}

func NewJobRunConsumer(brokers []string,topic string, groupID string,uc *usecase.UpdateJobStatusUsecase, l *zap.Logger) *JobRunConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic: topic,
		GroupID: groupID,
	})

	return &JobRunConsumer{
		reader: reader,
		usecase: uc,
		logger: l,
	}
}

func (c *JobRunConsumer) Start(ctx context.Context)error {
	c.logger.Info("job_run consumer started")
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("job_run conusmer shutting down")
			return nil 
		default:
			msg,err := c.reader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err,context.Canceled) {
					return nil 
				}
				c.logger.Error("kafka read failed",zap.Error(err))
				continue
			}

			var event domain.JobRunEvent
			if err := json.Unmarshal(msg.Value,&event); err != nil {
				c.logger.Error("invalid job_run message",zap.Error(err))
				continue 
			}

			if err := c.usecase.Handle(ctx, event); err != nil {
				c.logger.Error(
					"failed to update job status",
					zap.String("job_id",event.JobID),
					zap.Error(err),
				)
			}
		}	
	}
}

func (c *JobRunConsumer) Close() error {
	return c.reader.Close()
}