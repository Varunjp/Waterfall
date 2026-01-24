package domain

import "time"

type Job struct {
	JobID     	string
	AppID     	string
	Type      	string
	Payload   	string
	Status    	JobStatus
	CreatedAt 	time.Time
	UpdateAt 	time.Time
	Retry 		int 
	MaxRetries  int 
	ScheduledAt time.Time
}
