package domain

type Job struct {
	JobID      string
	Type       string
	ScheduleAt string
	Payload    []byte
}