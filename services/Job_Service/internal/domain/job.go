package domain

import "time"

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
	Payload   string       `json:"payload"`
	EventType JobEventType `json:"event_type"`
	Timestamp time.Time	   `json:"timestamp"`
}