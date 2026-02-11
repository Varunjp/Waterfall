package interfaces

import (
	"consumer_service/internal/domain"
	"context"
)

type JobRepository interface {
	Insert(ctx context.Context, job domain.Job) error
	UpdatePayload(ctx context.Context, jobID, payload string) error
	UpdateStatus(ctx context.Context, jobID string, status domain.JobStatus) error
	RetryJob(ctx context.Context,jobID string, status domain.JobStatus, retry int) error
}