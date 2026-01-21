package repository

import (
	"context"
	"time"
	"watcher_service/internal/domain"
)

type JobRepository interface {
	FetchDueJobs(ctx context.Context,now, until time.Time) ([]domain.Job,error)
	MarkQueued(ctx context.Context, jobID string)error 
}