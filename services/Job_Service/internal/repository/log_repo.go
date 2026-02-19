package repository

import (
	"context"
	"job_service/internal/domain"
)

type JobLogRepository interface {
	GetByJobID(ctx context.Context, jobID, appID string) ([]domain.JobLog, error)
}