package usecase

import (
	"consumer_service/internal/domain"
	"consumer_service/internal/repository/interfaces"
	"context"

	"go.uber.org/zap"
)

type ConsumeJobUsecase struct {
	repo interfaces.JobRepository
	logger *zap.Logger
}

func NewConsumeJobUsecase(r interfaces.JobRepository, l *zap.Logger) *ConsumeJobUsecase {
	return &ConsumeJobUsecase{repo: r,logger: l}
}

func (uc *ConsumeJobUsecase) Handle(ctx context.Context, event domain.JobEvent) error {
	
	switch event.EventType {
	case domain.JobCreated:
		job := domain.Job{
			JobID: event.JobID,
			AppID: event.AppID,
			Type: event.Type,
			Payload: event.Payload,
			Status: domain.StatusScheduled,
			CreatedAt: event.Timestamp,
			UpdateAt: event.Timestamp,
			ScheduleAt: event.Timestamp,
		}
		return uc.repo.Insert(ctx,job)
	case domain.JobUpdated:
		return uc.repo.UpdatePayload(ctx,event.JobID,event.Payload)
	case domain.JobCanceled:
		return uc.repo.UpdateStatus(ctx,event.JobID,domain.StatusCanceled)
	case domain.JobFailed:
		return uc.repo.UpdateStatus(ctx,event.JobID,domain.StatusFailed)
	case domain.JobComplete:
		return uc.repo.UpdateStatus(ctx,event.JobID,domain.StatusSuccess)
	case domain.JobRetry:
		return uc.repo.RetryJob(ctx,event.JobID,domain.StatusScheduled,event.Retry)
	}

	return nil 
}