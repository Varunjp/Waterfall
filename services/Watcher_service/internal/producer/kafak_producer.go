package producer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type kafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(broker []string, topic string) Producer {
	return &kafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(broker...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchSize:    100,
			BatchTimeout: 5 * time.Millisecond,
			RequiredAcks: kafka.RequireOne,
		},
	}
}

func (p *kafkaProducer) Publish(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: data,
	})
}

func (p *kafkaProducer) Close() error {
	return p.writer.Close()
}
