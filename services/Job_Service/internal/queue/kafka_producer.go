package queue

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type kafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) Producer {
	return &kafkaProducer{
		writer: &kafka.Writer{
			Addr: kafka.TCP(brokers...),
			Topic: topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (k *kafkaProducer) Publish(ctx context.Context,job JobEvent) error {
	bytes,err := json.Marshal(job)
	if err != nil {
		return err 
	}

	return k.writer.WriteMessages(ctx, kafka.Message{
		Value: bytes,
	})
}