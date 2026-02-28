package queue

import (
	"context"
	"time"
)

type JobEventType string

const (
	JobCreated  JobEventType = "JOB_CREATED"
	JobUpdated  JobEventType = "JOB_UPDATED"
	JobCanceled JobEventType = "JOB_CANCELED"
	JobRetry    JobEventType = "MANUAL_RETRY"
)

type JobEvent struct {
	JobID     string       `json:"job_id"`
	AppID     string       `json:"app_id"`
	Type      string       `json:"type"`
	Status 	  string 	   `json:"status"`
	Payload   string       `json:"payload"`
	EventType JobEventType `json:"event_type"`
	Timestamp time.Time    `json:"timestamp"`
	ManualRetry int 	   `json:"manual_retry"`
}

type Producer interface {
	Publish (ctx context.Context, job JobEvent) error 
}