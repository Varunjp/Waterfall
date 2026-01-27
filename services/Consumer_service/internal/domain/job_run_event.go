package domain

import "time"

type JobRunEvent struct {
	JobID      string  `json:"job_id"`
	Status     string  `json:"status"`
	Error      *string `json:"error,omitempty"`
	FinishedAt *time.Time `json:"finised_at,omitempty"`
}

