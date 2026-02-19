package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"watcher_service/internal/domain"
	"watcher_service/internal/usecase"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type KafkaConsumer struct {
	reader  *kafka.Reader
	usecase *usecase.ConsumeJobUsecase
	logger  *zap.Logger
}

func NewKafkaConsumer(
	brokers []string,
	topic string,
	groupID string,
	uc *usecase.ConsumeJobUsecase,
	l *zap.Logger,
) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &KafkaConsumer{
		reader:  reader,
		usecase: uc,
		logger:  l,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.logger.Info("job request consumer started")
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("job request consumer shutting down")
			return nil
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				c.logger.Info("kafka read failed", zap.Error(err))
				continue
			}

			var event domain.JobEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				c.logger.Error("invalid message", zap.Error(err))
				continue
			}

			if err := c.usecase.Handle(ctx, event); err != nil {
				c.logger.Error(
					"failed to process job",
					zap.String("job_id", event.JobID),
					zap.Error(err),
				)
			}
		}
	}
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
