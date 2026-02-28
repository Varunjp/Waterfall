package domain

import "time"

type RunningJob struct {
	JobID     string
	AppID     string
	JobType   string
	WorkerID  string
	StartedAt time.Time
	LastBeat  time.Time
	Retry     int
	MaxRetry  int
	ManualRetry int 
}