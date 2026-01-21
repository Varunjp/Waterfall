package usecase

import (
	"context"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/producer"

	"go.uber.org/zap"
)

type ScheduleJobUsecase struct {
	producer producer.Producer
	logger *zap.Logger
}

func NewSchedulerJobUsecase(p producer.Producer,l *zap.Logger) *ScheduleJobUsecase {
	return &ScheduleJobUsecase{producer: p, logger: l}
}

func (uc *ScheduleJobUsecase) Dispatch(ctx context.Context, job domain.Job) error {
	workerJob := domain.WorkerJob(job)

	err := uc.producer.Publish(ctx, job.JobID,workerJob)
	if err != nil {
		return err 
	}

	uc.logger.Info("job dispatched to worker", 
		zap.String("job_id",job.JobID),
	)
	return nil 
}