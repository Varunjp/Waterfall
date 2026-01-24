package domain

type JobStatus string

const (
	JobQueued  JobStatus = "QUEUED"
	JobRunning JobStatus = "RUNNING"
	JobFailed  JobStatus = "FAILED"
	JobSuccess JobStatus = "SUCCESS"
)