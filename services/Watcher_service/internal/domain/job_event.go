package domain

import "time"

type JobEventType string

const (
	JobCreated  JobEventType = "JOB_CREATED"
	JobUpdated  JobEventType = "JOB_UPDATED"
	JobCanceled JobEventType = "JOB_CANCELED"
	JobFailed   JobEventType = "JOB_FAILED"
	JobComplete JobEventType = "JOB_COMPLETED"
	JobRetry    JobEventType = "JOB_RETRY"
	ManualRetry JobEventType = "MANUAL_RETRY"
)

type JobEvent struct {
	JobID     	string       `json:"job_id"`
	AppID     	string       `json:"app_id"`
	Type      	string       `json:"type"`
	Payload   	string       `json:"payload"`
	EventType 	JobEventType `json:"event_type"`
	Timestamp 	time.Time    `json:"timestamp"`
	Retry     	int          `json:"retry"`
	ManualRetry int 		 `json:"manual_retry"`
}
