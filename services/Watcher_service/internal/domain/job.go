package domain

import "time"

type JobStatus string

const (
	StatusCreated   JobStatus = "CREATED"
	StatusCanceled  JobStatus = "CANCELED"
	StatusFailed    JobStatus = "FAILED"
	StatusSuccess   JobStatus = "COMPLETED"
	StatusScheduled JobStatus = "SCHEDULED"
	StatusQueued    JobStatus = "QUEUED"
)

type Job struct {
	JobID      string
	AppID      string
	Type       string
	Payload    string
	ScheduleAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Status     JobStatus
	Retry      int
	MaxRetries int
}

type QueueEvent struct {
	JobID      string `json:"job_id"`
	AppID      string `json:"app_id"`
	Type       string `json:"type"`
	Payload    string `json:"payload"`
	Retry      int    `json:"retry"`
	MaxRetries int    `json:"max_retries"`
}
