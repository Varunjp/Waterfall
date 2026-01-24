package consumer

type JobCreatedEvent struct {
	JobID      string `json:"job_id"`
	AppID      string `json:"app_id"`
	Type       string `json:"type"`
	Payload    []byte `json:"payload"`
	Retry      int    `json:"retry"`
	MaxRetries int    `json:"max_retries"`
}