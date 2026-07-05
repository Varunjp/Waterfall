package producer

import (
	"context"
	"encoding/json"
	"time"

	jobmetrics "job_service/internal/metrics"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	Publish(ctx context.Context, key string, value any) error
}

type kafkaProducer struct {
	writer  *kafka.Writer
	metrics *jobmetrics.JobMetrics
}

type testkafkaProducer struct {
	writer  *kafka.Writer
	metrics *jobmetrics.JobMetrics
}

func NewKafkaProducer(brokers []string, topic string, m *jobmetrics.JobMetrics) Producer {
	return &kafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchSize:    100,
			BatchTimeout: 5 * time.Millisecond,
			RequiredAcks: kafka.RequireOne,
		},
		metrics: m,
	}
}

func NewTestKafkaProducer(brokers []string, topic string, m *jobmetrics.JobMetrics) Producer {
	return &testkafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchSize:    100,
			BatchTimeout: 5 * time.Millisecond,
			RequiredAcks: kafka.RequireOne,
		},
		metrics: m,
	}
}

func (k *kafkaProducer) Publish(ctx context.Context, key string, value any) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		if k.metrics != nil {
			k.metrics.RecordKafkaPublish(false)
		}
		return err
	}

	err = k.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: bytes,
	})
	if k.metrics != nil {
		k.metrics.RecordKafkaPublish(err == nil)
	}
	return err
}

func (k *testkafkaProducer) Publish(ctx context.Context, key string, value any) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		if k.metrics != nil {
			k.metrics.RecordKafkaPublish(false)
		}
		return err
	}

	err = k.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: bytes,
	})
	if k.metrics != nil {
		k.metrics.RecordKafkaPublish(err == nil)
	}
	return err
}
