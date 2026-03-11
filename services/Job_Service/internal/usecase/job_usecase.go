package usecase

import (
	"context"
	"job_service/internal/domain"
	"job_service/internal/producer"
	redisRepo "job_service/internal/repository/redis"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type JobUseCase interface {
	Create(ctx context.Context, appID, jobType, payload string, scheduleAt *time.Time)(string, error)
	Update(ctx context.Context, jobID, payload string, scheduled_at time.Time,schduleModified bool) error 
	Cancel(ctx context.Context, jobID string) error 
}

type jobUsecase struct {
	producer producer.Producer
	logger *zap.Logger
	redis *redisRepo.RedisRepo
}

func NewJobUsecase(p producer.Producer,l *zap.Logger,r *redisRepo.RedisRepo) JobUseCase {
	return &jobUsecase{producer: p, logger: l,redis: r}
}

func (u *jobUsecase) Create(ctx context.Context, appID, jobType, payload string, scheduleAt *time.Time)(string, error) {

	err := u.redis.CheckQuota(ctx,appID)
	if err != nil {
		return "",err 
	}

	jobID := uuid.NewString()

	event := domain.JobEvent {
		JobID: jobID,
		AppID: appID,
		Type: jobType,
		Payload: payload,
		EventType: domain.JobCreated,
	}

	if scheduleAt != nil && scheduleAt.After(time.Now()) {
		event.Timestamp = *scheduleAt
	}else {
		event.Timestamp = time.Now()
	}
	
	err = u.producer.Publish(ctx,jobID,event)
	if err != nil {
		return "",err 
	}

	u.logger.Info("job created",zap.String("job_id",jobID))
	return jobID,nil 
}

func (u *jobUsecase) Update(ctx context.Context, jobID, payload string,scheduled_at time.Time,schduleModified bool) error {

	event := domain.JobEvent {
		JobID: jobID,
		Payload: payload,
		EventType: domain.JobUpdated,
		Timestamp: scheduled_at,
		ScheduleModifed: schduleModified,
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