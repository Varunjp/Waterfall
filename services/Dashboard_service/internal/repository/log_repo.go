package repository

import (
	"context"
	"dashboard_service/internal/domain"
)

type JobLogRepository interface {
	GetByJobID(ctx context.Context, jobID, appID string) ([]domain.JobLog, error)
}