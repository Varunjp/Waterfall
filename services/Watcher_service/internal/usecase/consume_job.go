package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"watcher_service/internal/domain"
	"watcher_service/internal/repository"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

type ConsumeJobUsecase struct {
	repo   repository.JobRepository
	adminRepo repository.AdminRepository
	redis  *redis.Client
	logger *zap.Logger
}

func NewConsumeJobUsecase(r repository.JobRepository, adrepo repository.AdminRepository, l *zap.Logger,rd *redis.Client) *ConsumeJobUsecase {
	return &ConsumeJobUsecase{repo: r, adminRepo:  adrepo,logger: l,redis: rd}
}

func (uc *ConsumeJobUsecase) Handle(ctx context.Context, event domain.JobEvent) error {

	switch event.EventType {
	case domain.JobCreated:
		job := domain.Job{
			JobID:      event.JobID,
			AppID:      event.AppID,
			Type:       event.Type,
			Payload:    event.Payload,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			ScheduleAt: event.Timestamp,
			ManualRetry: 0,
		}

		if event.Timestamp.After(time.Now()) {
			job.Status = domain.StatusPending
		}else {
			job.Status = domain.StatusScheduled
		}

		err := uc.repo.Insert(ctx, job)

		if err != nil {
			return err 
		}

		err = uc.adminRepo.UpdateUsageIncr(ctx,event.AppID)
		
		if err != nil {
			return err 
		}
		return nil 
	case domain.JobUpdated:
		updated,err := uc.repo.UpdatePayload(ctx, event.JobID, event.Payload,event.Timestamp,event.ScheduleModifed)
		if err != nil {
			return err 
		}
		if !updated {
			return errors.New("job cannot updated (already running or finished)")
		}
		return nil 
	case domain.JobCanceled:
		updated,err := uc.repo.CancelJob(ctx, event.JobID)
		if err != nil {
			return err 
		}
		if !updated {
			return errors.New("job cannot be cancelled (already running or finished)")
		}

		err = uc.adminRepo.UpdateUsageDecr(ctx,event.AppID)

		if err != nil {
			return err 
		}

		key := fmt.Sprintf("usage:%s:%s",event.AppID,time.Now().Format("2006-01"))
		err = uc.redis.Decr(key).Err()
		if err != nil {
			return err 
		}
		
		return nil 
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
