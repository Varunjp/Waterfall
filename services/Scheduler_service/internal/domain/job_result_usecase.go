package domain

import "context"

type JobResultInput struct {
	JobID        string
	AppID        string
	Status       string
	Retry        int
	ManualRetry  int 
	ErrorMessage string
}

type JobResultUsecase interface {
	ProcessJobResult(ctx context.Context,input JobResultInput) error 
}