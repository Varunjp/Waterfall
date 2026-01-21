package handler

import (
	"context"
	"job_service/internal/proto/jobpb"
	"job_service/internal/usecase"
)

type JobHandler struct {
	jobpb.UnimplementedJobServiceServer
	uc usecase.JobUseCase
}

func NewJobHandler(uc usecase.JobUseCase) *JobHandler {
	return &JobHandler{uc: uc}
}

func (h *JobHandler) CreateJob(ctx context.Context, r *jobpb.CreateJobRequest)(*jobpb.JobResponse, error) {
	jobID, err := h.uc.Create(ctx, r.AppId, r.Type, r.Payload)
	if err != nil {
		return nil, err 
	}
	return &jobpb.JobResponse{JobId: jobID,Status: "CREATED"},nil 
}

func (h *JobHandler) UpdateJob(ctx context.Context, r *jobpb.UpdateJobRequest) (*jobpb.JobResponse,error) {
	err := h.uc.Update(ctx,r.JobId,r.Payload)
	if err != nil {
		return nil,err 
	}
	return &jobpb.JobResponse{JobId: r.JobId, Status: "UPDATED"},nil 
}

func (h *JobHandler) CancelJob(ctx context.Context, r *jobpb.CancelJobRequest)(*jobpb.JobResponse, error) {
	err := h.uc.Cancel(ctx, r.JobId)
	if err != nil {
		return nil,err 
	}
	return &jobpb.JobResponse{JobId: r.JobId,Status: "CANCELED"},nil 
}