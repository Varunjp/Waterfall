package domain

import (
	"errors"
	"time"
)

type Worker struct {
	WorkerID     string
	AppID        string
	Capabilities []string
	Concurrency  int
	ActiveJobs   int
	LastSeen     time.Time
}

var ErrWorkerDisconnected error = errors.New("Worker disconnected")
var ErrWorkerQueueFull error = errors.New("worker queue full")
