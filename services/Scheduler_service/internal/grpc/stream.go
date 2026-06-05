package grpcserver

import (
	"context"
	"scheduler_service/internal/domain"
)

type GrpcDispatch interface {
	DispatchJob(ctx context.Context, job domain.Job, worker *WorkerConnection) error
}
