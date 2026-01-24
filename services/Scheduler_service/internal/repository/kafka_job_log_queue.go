package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type kafkaJobLogQueue struct {
	writer *kafka.Writer
}

func NewKafkaJobLogQueue(broker, topic string) *kafkaJobLogQueue {
	return &kafkaJobLogQueue{
		writer: &kafka.Writer{
			Addr: kafka.TCP(broker),
			Topic: topic,
		},
	}
}

func (k *kafkaJobLogQueue) Push(jobID, workerID ,status, error string, attempt int, logtime time.Time) error {
	log := map[string]any {
		"job_id": jobID,
		"worker_id":workerID,
		"status": status,
		"attempt": attempt,
		"error": error,
		"log_time": logtime,
	}

	data, _ := json.Marshal(log)
	return k.writer.WriteMessages(context.Background(),kafka.Message{
		Key: []byte(jobID),
		Value: data,
	})
}