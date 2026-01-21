package domain

type Job struct {
	JobID   string `json:"job_id"`
	AppID   string `json:"app_id"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type WorkerJob struct {
	JobID   string `json:"job_id"`
	AppID   string `json:"app_id"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
}
