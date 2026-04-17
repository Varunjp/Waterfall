package domain

import "time"

type WorkerRuntimeStatus string

const (
	WorkerStatusOnline  WorkerRuntimeStatus = "ONLINE"
	WorkerStatusBusy    WorkerRuntimeStatus = "BUSY"
	WorkerStatusStale   WorkerRuntimeStatus = "STALE"
	WorkerStatusOffline WorkerRuntimeStatus = "OFFLINE"
)

type WorkerSnapshot struct {
	WorkerID       string
	AppID          string
	JobTypes       []string
	ActiveJobs     int
	MaxConcurrency int
	LastSeen       time.Time
	Status         WorkerRuntimeStatus
}

type QueueSnapshot struct {
	JobType               string
	ReadyJobs             int64
	RunningJobs           int64
	RegisteredWorkers     int64
	BusyWorkers           int64
	OldestReadyAgeSeconds int64
}

type TenantRuntimeSnapshot struct {
	AppID            string
	GeneratedAt      time.Time
	TotalReadyJobs   int64
	TotalRunningJobs int64
	TotalWorkers     int64
	OnlineWorkers    int64
	BusyWorkers      int64
	Queues           []QueueSnapshot
	Workers          []WorkerSnapshot
}
