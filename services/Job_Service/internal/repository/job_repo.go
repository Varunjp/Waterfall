package repository

import (
	"context"
	"job_service/internal/domain"
)

type JobRepository interface {
	ListByApp(ctx context.Context,appID, status string,limit,offset int)([]domain.Job,error)
	ListFailed(ctx context.Context,limit,offset int)([]domain.Job,error)
	GetByID(ctx context.Context,jobID string)(*domain.Job,error)
}