package usecase

import (
	"context"
	"job_service/internal/domain"
	"job_service/internal/middleware"
	"job_service/internal/queue"
	"job_service/internal/repository"
	redisRepo "job_service/internal/repository/redis"
	"log"
	"time"
)

type DashboardUsecase struct {
	jobs repository.JobRepository
	logs repository.JobLogRepository
	queue queue.Producer
	redis *redisRepo.RedisRepo
}

func NewDashboardUsecase(j repository.JobRepository,l repository.JobLogRepository, q queue.Producer,r *redisRepo.RedisRepo) *DashboardUsecase {
	return &DashboardUsecase{jobs: j, logs: l, queue: q,redis: r}
}

func (uc *DashboardUsecase) ListJobs(ctx context.Context, status string, limit,offset int,startDate,endDate *time.Time)([]domain.Job,int,error) {
	
	appID,err := middleware.AppIDFromContext(ctx)
	if err != nil {
		log.Println(err.Error())
		return nil,0,err 
	}

	if !isValidTime(startDate) {
		startDate = nil 
	}

	if !isValidTime(endDate) {
		endDate = nil 
	}

	return uc.jobs.ListByApp(ctx,appID,status,limit,offset,startDate,endDate)
}

func (uc *DashboardUsecase) ListFailedJobs(ctx context.Context,limit,offset int)([]domain.Job,int,error) {
	return uc.jobs.ListFailed(ctx,limit,offset)
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

func (uc *DashboardUsecase) GetJobAdminLogs(ctx context.Context,jobID string)([]domain.JobLog,error) {
	return uc.logs.GetByJobIdAdmin(ctx,jobID)
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
	err = uc.redis.CheckQuota(ctx,job.AppID)
	if err != nil {
		return "",err 
	}

	err = uc.queue.Publish(ctx, queue.JobEvent{
		JobID: job.JobID,
		AppID: job.AppID,
		Type: job.Type,
		Status: string(domain.JobRetry),
		Payload: string(job.Payload),
		EventType: queue.JobRetry,
		ManualRetry: job.ManualRetry,
	})

	if err != nil {
		return "",err 
	}
	err = uc.redis.Incr(ctx,job.AppID)
	if err != nil {
		return "",err
	}
	return job.JobID,nil 
}

func isValidTime(t *time.Time) bool {
	if t == nil {
		return false
	}
	return !t.IsZero() && t.Unix() != 0
}