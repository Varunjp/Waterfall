package handler

import (
	"context"
	"job_service/internal/domain"
	"job_service/internal/proto/jobpb"
	"job_service/internal/usecase"
	"time"
)

type JobHandler struct {
	jobpb.UnimplementedJobServiceServer
	uc usecase.JobUseCase
	dc usecase.DashboardUsecase
}

func NewJobHandler(uc usecase.JobUseCase, dc usecase.DashboardUsecase) *JobHandler {
	return &JobHandler{uc: uc,dc : dc}
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

func (h *JobHandler) ListJobs(ctx context.Context,req *jobpb.ListJobsRequest)(*jobpb.ListJobsResponse,error) {

	jobs,err := h.dc.ListJobs(
		ctx,
		req.Status,
		int(req.Limit),
		int(req.Offset),
	)

	if err != nil {
		return nil,err 
	}

	return mapJobs(jobs),nil 
}

func (h *JobHandler) ListFailedJobs(ctx context.Context,req *jobpb.ListFailedJobsRequest)(*jobpb.ListJobsResponse,error) {
	jobs,err := h.dc.ListFailedJobs(
		ctx,
		int(req.Limit),
		int(req.Offset),
	)
	if err != nil {
		return nil,err 
	}

	return mapJobs(jobs),nil 
}

func (h *JobHandler) GetJobLogs(ctx context.Context, req *jobpb.GetJobLogsRequest)(*jobpb.GetJobLogsResponse,error) {
	logs,err := h.dc.GetJobLogs(ctx,req.JobId)
	if err != nil {
		return nil,err 
	}

	resp := &jobpb.GetJobLogsResponse{}
	for _,l := range logs {
		resp.Logs = append(resp.Logs, &jobpb.JobLog{
			Timestamp: l.Timestamp.Format(time.RFC3339),
			Level: l.Level,
			Message: l.Message,
		})
	}

	return resp,nil 
}

func (h *JobHandler) RetryJob(ctx context.Context, req *jobpb.RetryJobRequest)(*jobpb.RetryJobResponse,error) {
	id,err := h.dc.RetryJob(ctx,req.JobId)
	if err != nil {
		return nil,err 
	}

	return &jobpb.RetryJobResponse{NewJobId: id},nil 
}

func mapJobs(jobs []domain.Job) *jobpb.ListJobsResponse {
	resp := &jobpb.ListJobsResponse{}
	for _, j := range jobs {
		resp.Jobs = append(resp.Jobs, &jobpb.Job{
			JobId: j.JobID,
			AppId: j.AppID,
			Type: j.Type,
			Status: j.Status,
			Payload: string(j.Payload),
			Retry: int32(j.Retry),
			MaxRetry: int32(j.MaxRetry),
			CreatedAt: j.CreatedAt.Format(time.RFC3339),
			UpdatedAt: j.UpdatedAt.Format(time.RFC3339),
		})
	}

	return resp 
}