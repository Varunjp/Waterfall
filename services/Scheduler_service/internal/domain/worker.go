package domain

import "time"

type Worker struct {
	WorkerID   string 
	AppID      string 
	Capabilities []string 
	Concurrency int 
	ActiveJobs  int 
	LastSeen    time.Time 
}