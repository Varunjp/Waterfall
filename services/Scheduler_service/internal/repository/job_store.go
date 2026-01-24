package repository

import (
	"scheduler_service/internal/domain"
	"time"
)

type JobStore interface {
	SavePendingJob(domain.Job) error 
	PollJob(appID, jobType, workerId string)(*domain.Job,error)
	RefershRunning(jobID string)error 
	DeleteRunning(jobID string)error 
	Requeue(*domain.Job) error 
	FindStalled(timeout time.Duration)([]domain.RunningJob,error)
}