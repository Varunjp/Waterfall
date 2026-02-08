package grpcclient

import (
	"context"
	"dashboard_service/internal/domain"
	pb "dashboard_service/internal/transport/dashboardpb"
	"dashboard_service/internal/usecase"
	"time"
)

type Handler struct {
	pb.UnimplementedDashboardServiceServer
	uc *usecase.DashboardUsecase
}

func NewHandler(uc *usecase.DashboardUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) ListJobs(ctx context.Context,req *pb.ListJobsRequest)(*pb.ListJobsResponse,error) {

	jobs,err := h.uc.ListJobs(
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

func (h *Handler) ListFailedJobs(ctx context.Context,req *pb.ListFailedJobsRequest)(*pb.ListJobsResponse,error) {
	jobs,err := h.uc.ListFailedJobs(
		ctx,
		int(req.Limit),
		int(req.Offset),
	)
	if err != nil {
		return nil,err 
	}

	return mapJobs(jobs),nil 
}

func (h *Handler) GetJobLogs(ctx context.Context, req *pb.GetJobLogsRequest)(*pb.GetJobLogsResponse,error) {
	logs,err := h.uc.GetJobLogs(ctx,req.JobId)
	if err != nil {
		return nil,err 
	}

	resp := &pb.GetJobLogsResponse{}
	for _,l := range logs {
		resp.Logs = append(resp.Logs, &pb.JobLog{
			Timestamp: l.Timestamp.Format(time.RFC3339),
			Level: l.Level,
			Message: l.Message,
		})
	}

	return resp,nil 
}

func (h *Handler) RetryJob(ctx context.Context, req *pb.RetryJobRequest)(*pb.RetryJobResponse,error) {
	id,err := h.uc.RetryJob(ctx,req.JobId)
	if err != nil {
		return nil,err 
	}

	return &pb.RetryJobResponse{NewJobId: id},nil 
}

func mapJobs(jobs []domain.Job) *pb.ListJobsResponse {
	resp := &pb.ListJobsResponse{}
	for _, j := range jobs {
		resp.Jobs = append(resp.Jobs, &pb.Job{
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