package usecase

import (
	"context"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/metrics"
	"scheduler_service/internal/producer"
	"scheduler_service/internal/repository"

	"go.uber.org/zap"
)

type jobUsecase struct {
	adminRepo *repository.AdminRepo
	metrics *metrics.SchedulerMetrics
	producer *producer.KafkaProducer
	log *zap.Logger
	maxRetry int 
}

func NewJobResultProcess(a *repository.AdminRepo, m *metrics.SchedulerMetrics, p  *producer.KafkaProducer,l *zap.Logger,maxRetry int) *jobUsecase {
	return &jobUsecase{
		adminRepo: a,
		metrics: m,
		producer: p,
		log: l,
		maxRetry: maxRetry,
	}
}

func (u *jobUsecase) ProcessJobResult(ctx context.Context,input domain.JobResultInput) error {

	if input.Status == "FAILED" {
		if input.Retry < u.maxRetry {
			retry := "JOB_RETRY"
			event := map[string]any{
				"job_id": input.JobID,
				"app_id": input.AppID,
				"status": retry,
				"retry": input.Retry+1,
				"error": input.ErrorMessage,
			}
			if err := u.producer.Publish(ctx,event); err != nil {
				return err 
			}

			u.log.Info("job scheduled for retry",
				zap.String("job_id",input.JobID),
				zap.Int("retry",input.Retry+1),
			)

			return nil 
		}
	}

	event := map[string]any{
			"job_id": input.JobID,
			"app_id": input.AppID,
			"status": input.Status,
			"error": input.ErrorMessage,
		}

	if err := u.producer.Publish(ctx,event); err != nil {
		return err 
	}
	u.log.Info("job result processed",
		zap.String("job_id",input.JobID),
		zap.String("status",input.Status),
	)

	return nil 
}