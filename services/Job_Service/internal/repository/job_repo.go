package repository

import (
	"context"
	"job_service/internal/domain"
	"time"
)

type JobRepository interface {
	ListByApp(ctx context.Context,appID, status string,limit,offset int,startDate,endDate *time.Time)([]domain.Job,int,error)
	ListFailed(ctx context.Context,limit,offset int)([]domain.Job,int,error)
	GetByID(ctx context.Context,jobID string)(*domain.Job,error)
}