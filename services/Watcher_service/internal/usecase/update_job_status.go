package usecase

import (
	"context"
	"errors"
	"watcher_service/internal/domain"
	"watcher_service/internal/repository"
)

type UpdateJobStatusUsecase struct {
	jobRepo repository.JobRepository
}

func NewUpdateJobStatusUsecase(repo repository.JobRepository) *UpdateJobStatusUsecase {
	return &UpdateJobStatusUsecase{jobRepo: repo}
}

func (uc *UpdateJobStatusUsecase) Handle(ctx context.Context, event domain.JobRunEvent) error {
	if event.JobID == "" {
		return errors.New("job_id missing")
	}

	if event.Status == string(domain.JobRetry) {
		return uc.jobRepo.RetryJob(ctx, event.JobID, domain.StatusScheduled, event.Retry)
	}

	return uc.jobRepo.UpdateStatus(ctx, event.JobID, domain.JobStatus(event.Status))
}
