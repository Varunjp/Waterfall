package usecase

import (
	"context"
	"time"
	"watcher_service/internal/domain"
	"watcher_service/internal/producer"
	"watcher_service/internal/repository"

	"go.uber.org/zap"
)

type WatchJobsUsecase struct {
	repo     repository.JobRepository
	producer producer.Producer
	jobStatusproducer producer.Producer
	logger   *zap.Logger
}

func NewWatchJobsUsecase(
	r repository.JobRepository,
	p producer.Producer,
	j producer.Producer,
	l *zap.Logger,
) *WatchJobsUsecase {
	return &WatchJobsUsecase{repo: r, producer: p,jobStatusproducer: j ,logger: l}
}

func (uc *WatchJobsUsecase) Run(ctx context.Context) error {
	now := time.Now()
	until := now.Add(5 * time.Minute)
	jobs, err := uc.repo.FetchDueJobs(ctx, now, until)
	if err != nil {
		uc.logger.Error("job fetch from db failed",
			zap.Error(err),
		)
		return err
	}

	for _, job := range jobs {
		event := domain.QueueEvent{
			JobID:      job.JobID,
			AppID:      job.AppID,
			Type:       job.Type,
			Payload:    job.Payload,
			Retry:      job.Retry,
			MaxRetries: job.MaxRetries,
			ManualRetry: job.ManualRetry,
		}

		if err := uc.producer.Publish(ctx, job.JobID, event); err != nil {
			uc.logger.Error("kafka publish failed",
				zap.String("job_id", job.JobID),
				zap.Error(err),
			)
			continue
		}

		err := uc.repo.MarkQueued(ctx, job.JobID)
		if err != nil {
			uc.logger.Error("job update failed",
				zap.String("job_id", job.JobID),
				zap.Error(err),
			)
		}
		uc.logger.Info("job queued", zap.String("job_id", job.JobID))
	}

	runJobs,err := uc.repo.RunningJobs(ctx)
	if err != nil {
		uc.logger.Error("job fetch from db failed",
			zap.Error(err),
		)
		return err 
	}

	for _, job := range runJobs {
		event := domain.QueueEvent{
			JobID:      job.JobID,
			AppID:      job.AppID,
			Type:       job.Type,
			Payload:    job.Payload,
			Retry:      job.Retry,
			MaxRetries: job.MaxRetries,
			ManualRetry: job.ManualRetry,
		}

		if err := uc.jobStatusproducer.Publish(ctx, job.JobID, event); err != nil {
			uc.logger.Error("kafka publish failed",
				zap.String("job_id", job.JobID),
				zap.Error(err),
			)
			continue
		}
		uc.logger.Info("job queued", zap.String("job_id", job.JobID))
	}

	return nil
}
