package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer {
		writer : &kafka.Writer{
			Addr: kafka.TCP(brokers...),
		},
	}
}

func (p *Producer) Publish(topic string, key string, value []byte) error {
	return p.writer.WriteMessages(context.Background(),
		kafka.Message{
			Topic: topic,
			Key: []byte(key),
			Value: value,
		},
	)
}