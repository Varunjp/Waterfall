package consumer

type JobCreatedEvent struct {
	JobID       string `json:"job_id"`
	AppID       string `json:"app_id"`
	Type        string `json:"type"`
	Payload     string `json:"payload"`
	Retry       int    `json:"retry"`
	MaxRetries  int    `json:"max_retries"`
	ManualRetry int    `json:"manual_retry"`
}