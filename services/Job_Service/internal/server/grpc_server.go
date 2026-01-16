package server

import (
	"context"
	"job_service/internal/domain"
	jobpb "job_service/internal/proto"
	"job_service/internal/usecase"

	"go.opentelemetry.io/otel"
)

type GRPCServer struct {
	jobpb.UnimplementedJobServiceServer
	uc *usecase.JobUsecase
}

func New(uc *usecase.JobUsecase) *GRPCServer {
	return &GRPCServer{uc: uc}
}

func (s *GRPCServer) CreateJob(ctx context.Context, r *jobpb.CreateJobRequest) (*jobpb.JobResponse,error) {
	tr := otel.Tracer("job-service")
	ctx,span := tr.Start(ctx,"CreateJob")
	defer span.End()

	job := domain.Job {
		JobID: r.IdempotenceyKey,
		Type: r.JobType,
		Payload: r.Payload,
		ScheduleAt: r.ScheduleAt,
	}

	err := s.uc.CreateJob(ctx,job,r.IdempotenceyKey)
	if err != nil {
		return  nil,err 
	}

	return &jobpb.JobResponse{JobId: job.JobID,Status: "CREATED"},nil 
}

func (s *GRPCServer) CancleJob(ctx context.Context, r *jobpb.CancelJobRequest) (*jobpb.JobResponse,error) {
	err := s.uc.CancelJob(ctx,r.JobId)

	if err != nil {
		return nil,err 
	}

	return &jobpb.JobResponse{JobId: r.JobId, Status: "CANCELED"},nil 
}

func (s *GRPCServer) TriggerJobNow(ctx context.Context, r *jobpb.TriggerNowRequest) (*jobpb.JobResponse,error) {
	err := s.uc.TriggerNow(ctx,r.JobId)
	if err != nil {
		return nil,err 
	}

	return &jobpb.JobResponse{JobId: r.JobId, Status: "TRIGGERED"},nil 
}