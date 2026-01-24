package repository

import "scheduler_service/internal/domain"

type WorkerRegistry interface {
	Register(worker domain.Worker) error 
	Heartbeat(workerID string) error 
	Acquire(appID, jobType string)(*domain.Worker,error)
	Release(workerID string) error 
}