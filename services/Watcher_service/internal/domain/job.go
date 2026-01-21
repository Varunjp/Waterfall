package domain

import "time"

type JobStatus string

const (
	StatusScheduled JobStatus = "SCHEDULED"
	StatusQueued    JobStatus = "QUEUED"
)

type Job struct {
	JobID      string
	AppID      string
	Type       string
	Payload    string
	ScheduleAt time.Time
	Status 		JobStatus
}

type QueueEvent struct {
	JobID 		string  `json:"job_id"`
	AppID 		string  `json:"app_id"`
	Type 		string  `json:"type"`
	Payload     string  `json:"payload"`
}