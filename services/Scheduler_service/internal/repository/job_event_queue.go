package repository

import "scheduler_service/internal/domain"

type JobEventQueue interface {
	JobAssigned(job domain.Job) error 
	JobRunning(jobID, workerID string) error 
	JobSucceeded(jobID string) error 
	JobFailed(jobID, reason string) error 
}