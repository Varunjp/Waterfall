package producer

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	Publish(ctx context.Context, key string, value any) error
}

type kafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string,topic string) Producer {
	return &kafkaProducer {
		writer: &kafka.Writer{
			Addr: kafka.TCP(brokers...),
			Topic: topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (k *kafkaProducer) Publish(ctx context.Context,key string,value any)error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err 
	}

	return k.writer.WriteMessages(ctx, kafka.Message{
		Key: []byte(key),
		Value: bytes,
	})
}