package domain

import "time"

type Job struct {
	JobID     	string `json:"job_id"`
	AppID     	string `json:"app_id"`
	Type      	string `json:"job_type"`
	Payload   	string `json:"payload"`
	Status    	JobStatus
	CreatedAt 	time.Time
	UpdateAt 	time.Time
	Retry 		int 
	MaxRetries  int `json:"max_retries"`
	ScheduledAt time.Time
}
