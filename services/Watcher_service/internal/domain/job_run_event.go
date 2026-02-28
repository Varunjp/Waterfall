package domain

import "time"

type JobRunEvent struct {
	JobID      string     `json:"job_id"`
	Status     string     `json:"status"`
	Retry      int        `json:"retry"`
	Error      *string    `json:"error,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	NextRun    *time.Time `json:"next_run,omitempty"`
}
