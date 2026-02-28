package repository

import (
	"context"
	"time"
	"watcher_service/internal/domain"
)

type JobRepository interface {
	FetchDueJobs(ctx context.Context, now, until time.Time) ([]domain.Job, error)
	MarkQueued(ctx context.Context, jobID string) error
	Insert(ctx context.Context, job domain.Job) error
	UpdatePayload(ctx context.Context, jobID, payload string) error
	UpdateStatus(ctx context.Context, jobID string, status domain.JobStatus) error
	RetryJob(ctx context.Context, jobID string, status domain.JobStatus, retry int,nextRun time.Time) error
	JobLog(ctx context.Context,jobEvent domain.JobRunEvent) error
	RunningJobs(ctx context.Context)([]domain.Job,error)
	JobManualRetry(ctx context.Context,jobID string, status domain.JobStatus) error 
}
