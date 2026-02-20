package usecase

import (
	"context"
	"job_service/internal/domain"
	"job_service/internal/middleware"
	"job_service/internal/queue"
	"job_service/internal/repository"
	"log"

	"github.com/google/uuid"
)

type DashboardUsecase struct {
	jobs repository.JobRepository
	logs repository.JobLogRepository
	queue queue.Producer
}

func NewDashboardUsecase(j repository.JobRepository,l repository.JobLogRepository, q queue.Producer) *DashboardUsecase {
	return &DashboardUsecase{jobs: j, logs: l, queue: q}
}

func (uc *DashboardUsecase) ListJobs(ctx context.Context, status string, limit,offset int)([]domain.Job,error) {
	
	appID,err := middleware.AppIDFromContext(ctx)
	if err != nil {
		log.Println(err.Error())
		return nil,err 
	}

	if status == "" {
		status = "COMPLETED"
	}

	return uc.jobs.ListByApp(ctx,appID,status,limit,offset)
}

func (uc *DashboardUsecase) ListFailedJobs(ctx context.Context,limit,offset int)([]domain.Job,error) {
	appID,err := middleware.AppIDFromContext(ctx)
	if err != nil {
		log.Println(err.Error())
		return nil,err 
	}
	return uc.jobs.ListFailed(ctx,appID,limit,offset)
}

func (uc *DashboardUsecase) GetJobLogs(
	ctx context.Context,jobID string,
)([]domain.JobLog,error) {
	appID,err := middleware.AppIDFromContext(ctx)
	if err != nil {
		log.Println(err.Error())
		return nil,err 
	}
	return uc.logs.GetByJobID(ctx,jobID,appID)
}

func (uc *DashboardUsecase) RetryJob(ctx context.Context,jobID string)(string,error) {

	role,err := middleware.RoleFromContext(ctx)

	if err != nil {
		return "",err
	}

	if role == "viewer" {
		return "",domain.ErrForbidden
	}

	job, err := uc.jobs.GetByID(ctx,jobID)
	if err != nil {
		return "",err 
	}

	if job.Retry >= job.MaxRetry {
		return "",domain.ErrMaxRetryExceeded
	}

	newJobID := uuid.NewString()

	err = uc.queue.Publish(ctx, queue.JobEvent{
		JobID: newJobID,
		AppID: job.AppID,
		Type: job.Type,
		Payload: string(job.Payload),
		EventType: queue.JobCreated,
	})

	if err != nil {
		return "",err 
	}

	return newJobID,nil 
}