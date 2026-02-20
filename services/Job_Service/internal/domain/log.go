package domain

import "time"

type JobLog struct {
	JobID     string
	Status     string
	ErrorMessage   string
	Timestamp time.Time
}