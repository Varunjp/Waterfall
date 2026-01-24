package repository

import (
	"context"
	"encoding/json"
	"scheduler_service/internal/domain"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaJobEventQueue struct {
	writer *kafka.Writer
}

func NewKafkaJobEventQueue(broker, topic string) *KafkaJobEventQueue {
	return &KafkaJobEventQueue{
		writer: &kafka.Writer{
			Addr: kafka.TCP(broker),
			Topic: topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (k *KafkaJobEventQueue) publish(key string, payload any) error {
	data,_ := json.Marshal(payload)
	return k.writer.WriteMessages(context.Background(),kafka.Message{
		Key: []byte(key),
		Value: data,
		Time: time.Now(),
	})
}

func (k *KafkaJobEventQueue)JobAssigned(job domain.Job) error {
	return k.publish(job.JobID, map[string]any{
		"job_id":job.JobID,
		"status": "ASSIGNED",
	})
} 

func (k *KafkaJobEventQueue) JobRunning(jobID, workerID string) error {
	return k.publish(jobID, map[string]any{
		"job_id":jobID,
		"worker_id":workerID,
		"status": "RUNNING",
	})
}

func (k *KafkaJobEventQueue) JobSucceeded(jobID string) error {
	return k.publish(jobID,map[string]any{
		"job_id":jobID,
		"status":"SUCCESS",
	})
}

func (k *KafkaJobEventQueue) JobFailed(jobID, reason string) error {
	return k.publish(jobID, map[string]any{
		"job_id":jobID,
		"status": "FAILED",
		"error": reason,
	})
}