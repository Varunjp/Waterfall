package domain

import "time"

type Job struct {
	JobID     	string
	AppID     	string
	Type      	string
	Payload   	[]byte
	Status    	string
	Retry     	int
	MaxRetry  	int
	ManualRetry int 
	CreatedAt 	time.Time
	UpdatedAt 	time.Time
	ScheduledAt time.Time
}

type JobStats struct {
	TotalJobs        int32
    TotalSuccessJobs int32
    TotalFailedJobs  int32
}

type MetricBucket struct {
	TS        string
	Created   int32
	Completed int32
	Failed    int32
}