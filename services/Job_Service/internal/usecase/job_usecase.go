package usecase

import (
	"context"
	"job_service/internal/domain"
	"job_service/internal/producer"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type JobUseCase interface {
	Create(ctx context.Context, appID, jobType, payload string)(string, error)
	Update(ctx context.Context, jobID, payload string) error 
	Cancel(ctx context.Context, jobID string) error 
}

type jobUsecase struct {
	producer producer.Producer
	logger *zap.Logger
}

func NewJobUsecase(p producer.Producer,l *zap.Logger) JobUseCase {
	return &jobUsecase{producer: p, logger: l}
}

func (u *jobUsecase) Create(ctx context.Context, appID, jobType, payload string)(string, error) {
	jobID := uuid.NewString()

	event := domain.JobEvent {
		JobID: jobID,
		AppID: appID,
		Type: jobType,
		Payload: payload,
		EventType: domain.JobCreated,
		Timestamp: time.Now(),
	}

	err := u.producer.Publish(ctx,jobID,event)
	if err != nil {
		return "",err 
	}

	u.logger.Info("job created",zap.String("job_id",jobID))
	return jobID,nil 
}

func (u *jobUsecase) Update(ctx context.Context, jobID, payload string) error {
	event := domain.JobEvent {
		JobID: jobID,
		Payload: payload,
		EventType: domain.JobUpdated,
		Timestamp: time.Now(),
	}

	return u.producer.Publish(ctx,jobID,event)
}

func (u *jobUsecase) Cancel(ctx context.Context, jobID string) error {
	event := domain.JobEvent {
		JobID: jobID,
		EventType: domain.JobCanceled,
		Timestamp: time.Now(),
	}
	// need to add redis call 
	return u.producer.Publish(ctx, jobID, event)
}