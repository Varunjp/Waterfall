package domain

import "time"

type Job struct {
	JobID     string
	AppID     string
	Type      string
	Payload   []byte
	Status    string
	Retry     int
	MaxRetry  int
	ManualRetry int 
	CreatedAt time.Time
	UpdatedAt time.Time
}