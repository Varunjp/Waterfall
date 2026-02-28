package usecase

import (
	"context"
	"watcher_service/internal/domain"
	"watcher_service/internal/repository"

	"go.uber.org/zap"
)

type ConsumeJobUsecase struct {
	repo   repository.JobRepository
	logger *zap.Logger
}

func NewConsumeJobUsecase(r repository.JobRepository, l *zap.Logger) *ConsumeJobUsecase {
	return &ConsumeJobUsecase{repo: r, logger: l}
}

func (uc *ConsumeJobUsecase) Handle(ctx context.Context, event domain.JobEvent) error {

	switch event.EventType {
	case domain.JobCreated:
		job := domain.Job{
			JobID:      event.JobID,
			AppID:      event.AppID,
			Type:       event.Type,
			Payload:    event.Payload,
			Status:     domain.StatusScheduled,
			CreatedAt:  event.Timestamp,
			UpdatedAt:  event.Timestamp,
			ScheduleAt: event.Timestamp,
			ManualRetry: 0,
		}
		return uc.repo.Insert(ctx, job)
	case domain.JobUpdated:
		return uc.repo.UpdatePayload(ctx, event.JobID, event.Payload)
	case domain.JobCanceled:
		return uc.repo.UpdateStatus(ctx, event.JobID, domain.StatusCanceled)
	case domain.JobFailed:
		return uc.repo.UpdateStatus(ctx, event.JobID, domain.StatusFailed)
	case domain.JobComplete:
		return uc.repo.UpdateStatus(ctx, event.JobID, domain.StatusSuccess)
	case domain.JobRetry:
		return uc.repo.RetryJob(ctx, event.JobID, domain.StatusScheduled, event.Retry,event.Timestamp)
	case domain.ManualRetry:
		return uc.repo.JobManualRetry(ctx,event.JobID,domain.StatusManualRetry)
	}

	return nil
}
