package usecase

import (
	"context"
	"errors"
	"time"
	"watcher_service/internal/domain"
	"watcher_service/internal/repository"

	"go.uber.org/zap"
)

type ConsumeTestJobUsecase struct {
	repo      repository.JobRepository
	adminRepo repository.AdminRepository
	logger    *zap.Logger
}

func NewConsumetestJobUsecase(r repository.JobRepository, adrepo repository.AdminRepository, l *zap.Logger) *ConsumeTestJobUsecase {
	return &ConsumeTestJobUsecase{repo: r, adminRepo: adrepo, logger: l}
}

func (uc *ConsumeTestJobUsecase) Handle(ctx context.Context, event domain.JobEvent) error {

	switch event.EventType {
	case domain.JobTestCreated:
		now := time.Now().UTC()
		job := domain.Job{
			JobID:      event.JobID,
			AppID:      event.AppID,
			Type:       event.Type,
			Payload:    event.Payload,
			Status:     domain.StatusScheduled,
			CreatedAt:  now,
			UpdatedAt:  now,
			ScheduleAt: now,
		}

		err := uc.repo.InsertTest(ctx, job)

		if err != nil {
			return err
		}

		return nil

	case domain.JobUpdated:
		updated, err := uc.repo.UpdatePayload(ctx, event.JobID, event.Payload, event.Timestamp, event.ScheduleModifed)
		if err != nil {
			return err
		}
		if !updated {
			return errors.New("job cannot updated (already running or finished)")
		}
		return nil
	case domain.JobCanceled:
		updated, err := uc.repo.CancelJob(ctx, event.JobID)
		if err != nil {
			return err
		}
		if !updated {
			return errors.New("job cannot be cancelled (already running or finished)")
		}

		err = uc.adminRepo.UpdateUsageDecr(ctx, event.AppID)

		if err != nil {
			return err
		}

		return nil
	case domain.JobFailed:
		return uc.repo.UpdateStatus(ctx, event.JobID, domain.StatusFailed)
	case domain.JobComplete:
		return uc.repo.UpdateStatus(ctx, event.JobID, domain.StatusSuccess)
	case domain.JobRetry:
		return uc.repo.RetryJob(ctx, event.JobID, domain.StatusScheduled, event.Retry, event.Timestamp)
	case domain.ManualRetry:
		return uc.repo.JobManualRetry(ctx, event.JobID, domain.StatusManualRetry)
	}

	return nil
}
