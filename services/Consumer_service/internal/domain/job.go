package domain

import "time"

type JobStatus string

const (
	StatusCreated  JobStatus = "CREATED"
	StatusCanceled JobStatus = "CANCELED"
	StatusFailed   JobStatus = "FAILED"
	StatusSuccess  JobStatus = "SUCCESS"
)

type Job struct {
	JobID     	string
	AppID     	string
	Type      	string
	Payload   	string
	Status    	JobStatus
	CreatedAt 	time.Time
	UpdateAt 	time.Time
}

type JobEventType string 

const (
	JobCreated  JobEventType = "JOB_CREATED"
	JobUpdated  JobEventType = "JOB_UPDATED"
	JobCanceled JobEventType = "JOB_CANCELED"
)

type JobEvent struct {
	JobID     string       `json:"job_id"`
	AppID     string       `json:"app_id"`
	Type      string       `json:"type"`
	Payload   string       `json:"payload"`
	EventType JobEventType `json:"event_type"`
	Timestamp time.Time    `json:"timestamp"`
}