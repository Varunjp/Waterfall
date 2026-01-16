package usecase

import (
	"context"
	"job_service/internal/domain"
)

type KafkaProducer interface {
	Publish(topic string, key string, value []byte) error
}

type IdempotencyStore interface {
	IsProcessed(ctx context.Context,key string)(bool,error)
	MarkProcessed(ctx context.Context,key string) error 
}

type JobUsecase struct {
	kafka 		KafkaProducer
	idem 		IdempotencyStore
}

func NewJobUsecase(k KafkaProducer,i IdempotencyStore) *JobUsecase {
	return &JobUsecase{kafka: k,idem: i}
}

func (u *JobUsecase) CreateJob(ctx context.Context, job domain.Job, idemKey string) error {
	ok, _ := u.idem.IsProcessed(ctx,idemKey)
	if ok {
		return nil 
	}

	err := u.kafka.Publish("job-requests",job.JobID,job.Payload)
	if err != nil {
		return err 
	}

	return u.idem.MarkProcessed(ctx,idemKey)
}

func (u *JobUsecase) TriggerNow(ctx context.Context, jobID string) error {
	return u.kafka.Publish("job-trigger-now",jobID,[]byte("{}"))
}

func (u *JobUsecase) CancelJob(ctx context.Context, jobID string) error {
	return u.kafka.Publish("job-cancel",jobID,[]byte("{}"))
}