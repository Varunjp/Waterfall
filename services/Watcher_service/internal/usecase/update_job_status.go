package usecase

import (
	"context"
	"errors"
	"time"
	"watcher_service/internal/domain"
	"watcher_service/internal/repository"
)

type UpdateJobStatusUsecase struct {
	jobRepo repository.JobRepository
	adminRepo repository.AdminRepository
}

func NewUpdateJobStatusUsecase(repo repository.JobRepository,adrepo repository.AdminRepository) *UpdateJobStatusUsecase {
	return &UpdateJobStatusUsecase{jobRepo: repo,adminRepo: adrepo}
}

func (uc *UpdateJobStatusUsecase) Handle(ctx context.Context, event domain.JobRunEvent) error {
	if event.JobID == "" {
		return errors.New("job_id missing")
	}
	
	if event.Retry != 0  {
		err := uc.jobRepo.JobLog(ctx,event)
		if err != nil {
			return err 
		}
	}

	if event.Status == string(domain.JobRetry) {
		var nr time.Time 
		if event.NextRun != nil {
			nr = *event.NextRun
		}else {
			nr = time.Now()
		}
		return uc.jobRepo.RetryJob(ctx, event.JobID, domain.StatusScheduled, event.Retry,nr)
	}

	if event.Status == "COMPLETED" || event.Status == "FAILED" {
		err := uc.adminRepo.UpdateUsage(ctx,event.AppID)
		if err != nil {
			return err 
		}
	}

	return uc.jobRepo.UpdateStatus(ctx, event.JobID, domain.JobStatus(event.Status))
}
