package usecase

import (
	"context"
	"job_service/internal/domain"
	jobmetrics "job_service/internal/metrics"
	"job_service/internal/producer"
	redisRepo "job_service/internal/repository/redis"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type JobUseCase interface {
	Create(ctx context.Context, appID, jobType, payload string, scheduleAt *time.Time) (string, error)
	CreateTest(ctx context.Context, appID, jobType, payload string) (string, error)
	Update(ctx context.Context, jobID, payload string, scheduled_at time.Time, schduleModified bool) error
	Cancel(ctx context.Context, jobID string) error
}

type jobUsecase struct {
	producer producer.Producer
	test     producer.Producer
	logger   *zap.Logger
	redis    *redisRepo.RedisRepo
	metrics  *jobmetrics.JobMetrics
}

func NewJobUsecase(p producer.Producer, t producer.Producer, l *zap.Logger, r *redisRepo.RedisRepo, m *jobmetrics.JobMetrics) JobUseCase {
	return &jobUsecase{producer: p, test: t, logger: l, redis: r, metrics: m}
}

func (u *jobUsecase) Create(ctx context.Context, appID, jobType, payload string, scheduleAt *time.Time) (string, error) {
	start := time.Now()
	defer func() {
		if u.metrics != nil {
			u.metrics.ObserveCreateDuration(time.Since(start))
		}
	}()

	err := u.redis.CheckQuota(ctx, appID)
	if err != nil {
		if u.metrics != nil {
			u.metrics.RecordJobCreation(len(payload), false)
		}
		return "", err
	}

	jobID := uuid.NewString()

	event := domain.JobEvent{
		JobID:     jobID,
		AppID:     appID,
		Type:      jobType,
		Payload:   payload,
		EventType: domain.JobCreated,
	}

	now := time.Now().UTC()
	if scheduleAt != nil {
		scheduledAtUTC := scheduleAt.UTC()
		if scheduledAtUTC.After(now) {
			event.Timestamp = scheduledAtUTC
		} else {
			event.Timestamp = now
		}
	} else {
		event.Timestamp = now
	}

	err = u.producer.Publish(ctx, jobID, event)
	if err != nil {
		if u.metrics != nil {
			u.metrics.RecordJobCreation(len(payload), false)
		}
		return "", err
	}

	err = u.redis.Incr(ctx, appID)

	if err != nil {
		if u.metrics != nil {
			u.metrics.RecordJobCreation(len(payload), false)
		}
		return "", err
	}

	if u.metrics != nil {
		u.metrics.RecordJobCreation(len(payload), true)
	}

	u.logger.Info("job created", zap.String("job_id", jobID))
	return jobID, nil
}

func (u *jobUsecase) CreateTest(ctx context.Context, appID, jobType, payload string) (string, error) {
	start := time.Now()
	defer func() {
		if u.metrics != nil {
			u.metrics.ObserveCreateDuration(time.Since(start))
		}
	}()

	jobID := uuid.NewString()

	event := domain.JobEvent{
		JobID:     jobID,
		AppID:     appID,
		Type:      jobType,
		Payload:   payload,
		EventType: domain.JobTestCreated,
	}

	err := u.test.Publish(ctx, jobID, event)
	if err != nil {
		if u.metrics != nil {
			u.metrics.RecordJobCreation(len(payload), false)
		}
		return "", err
	}

	if u.metrics != nil {
		u.metrics.RecordJobCreation(len(payload), true)
	}

	return jobID, nil
}

func (u *jobUsecase) Update(ctx context.Context, jobID, payload string, scheduled_at time.Time, schduleModified bool) error {

	event := domain.JobEvent{
		JobID:           jobID,
		Payload:         payload,
		EventType:       domain.JobUpdated,
		Timestamp:       scheduled_at.UTC(),
		ScheduleModifed: schduleModified,
	}

	return u.producer.Publish(ctx, jobID, event)
}

func (u *jobUsecase) Cancel(ctx context.Context, jobID string) error {
	event := domain.JobEvent{
		JobID:     jobID,
		EventType: domain.JobCanceled,
		Timestamp: time.Now().UTC(),
	}

	return u.producer.Publish(ctx, jobID, event)
}
