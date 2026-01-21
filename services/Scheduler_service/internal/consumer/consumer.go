package consumer

import (
	"context"
	"encoding/json"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/usecase"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	usecase *usecase.ScheduleJobUsecase
	logger *zap.Logger
}

func NewKafkaConsumer(
	brokers []string,
	topic string, 
	groupID string,
	uc *usecase.ScheduleJobUsecase,
	l *zap.Logger,
) *KafkaConsumer {

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic: topic,
		GroupID: groupID,
	})

	return &KafkaConsumer{
		reader: reader,
		usecase: uc,
		logger: l,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	for {
		msg,err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err 
		}

		var job domain.Job
		if err := json.Unmarshal(msg.Value, &job); err != nil {
			c.logger.Error("invalid job payload",zap.Error(err))
			continue 
		}

		if err := c.usecase.Dispatch(ctx,job); err != nil {
			c.logger.Error("dispatch failed",
				zap.String("job_id",job.JobID),
				zap.Error(err),
			)
		}
	}
}