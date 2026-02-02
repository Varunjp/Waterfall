package repository

import (
	"context"
	"scheduler_service/internal/domain"
	"time"
)

type JobStore interface {
	SavePendingJob(domain.Job) error 
	RefershRunning(jobID string)error 
	DeleteRunning(jobID string)error 
	Requeue(*domain.Job) error 
	FindStalled(timeout time.Duration)([]domain.RunningJob,error)
	ReadJob(ctx context.Context,appID,jobType,group,consumer string)(*domain.Job,string,error)
	PushJob(job domain.Job) error 
	EnsureGroup(appID,jobType,group string)error 
	AckJob(appID,jobType,group,messageID string)error 
}