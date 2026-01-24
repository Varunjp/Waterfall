package repository

import "time"

type JobLogQueue interface {
	Push(jobID, workerID, status, error string, attempt int, logtime time.Time) error
}