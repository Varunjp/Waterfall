package domain

import "time"

type JobLog struct {
	JobID     string
	Level     string
	Message   string
	Timestamp time.Time
}