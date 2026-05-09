package handler

import (
	"context"
	"job_service/internal/domain"
	"job_service/internal/proto/jobpb"
	"job_service/internal/usecase"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type JobHandler struct {
	jobpb.UnimplementedJobServiceServer
	uc usecase.JobUseCase
	dc usecase.DashboardUsecase
}

func NewJobHandler(uc usecase.JobUseCase, dc usecase.DashboardUsecase) *JobHandler {
	return &JobHandler{uc: uc, dc: dc}
}

func (h *JobHandler) CreateJob(ctx context.Context, r *jobpb.CreateJobRequest) (*jobpb.JobResponse, error) {
	jobID, err := h.uc.Create(ctx, r.AppId, r.Type, r.Payload, optionalTimestamp(r.ScheduleAt))
	if err != nil {
		return nil, err
	}
	return &jobpb.JobResponse{JobId: jobID, Status: "CREATED"}, nil
}

func (h *JobHandler) CreateTestJob(ctx context.Context,r *jobpb.CreateTestJobRequest)(*jobpb.JobResponse,error) {

	jobID,err := h.uc.CreateTest(ctx,r.AppId,r.Type,r.Payload)

	if err != nil {
		return nil,err 
	}
	return &jobpb.JobResponse{JobId: jobID,Status: "CREATED_TEST"},nil 
}	


func (h *JobHandler) UpdateJob(ctx context.Context, r *jobpb.UpdateJobRequest) (*jobpb.JobResponse, error) {

	var schedulTime time.Time
	scheModified := false
	if ts := optionalTimestamp(r.ScheduleAt); ts != nil {
		schedulTime = *ts
		scheModified = true
	}

	err := h.uc.Update(ctx, r.JobId, r.Payload, schedulTime, scheModified)
	if err != nil {
		return nil, err
	}
	return &jobpb.JobResponse{JobId: r.JobId, Status: "UPDATED"}, nil
}

func (h *JobHandler) CancelJob(ctx context.Context, r *jobpb.CancelJobRequest) (*jobpb.JobResponse, error) {
	err := h.uc.Cancel(ctx, r.JobId)
	if err != nil {
		return nil, err
	}
	return &jobpb.JobResponse{JobId: r.JobId, Status: "CANCELED"}, nil
}

func (h *JobHandler) ListJobs(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
	jobs, total, err := h.dc.ListJobs(
		ctx,
		req.Status,
		int(req.Limit),
		int(req.Offset),
		optionalTimestamp(req.StartDate),
		optionalTimestamp(req.EndDate),
	)

	if err != nil {
		return nil, err
	}

	return mapJobs(jobs, total, int(req.Limit), int(req.Offset)), nil
}

func (h *JobHandler) ListFailedJobs(ctx context.Context, req *jobpb.ListFailedJobsRequest) (*jobpb.ListJobsResponse, error) {
	jobs, total, err := h.dc.ListFailedJobs(
		ctx,
		int(req.Limit),
		int(req.Offset),
	)
	if err != nil {
		return nil, err
	}

	return mapJobs(jobs, total, int(req.Limit), int(req.Offset)), nil
}

func (h *JobHandler) GetJobLogs(ctx context.Context, req *jobpb.GetJobLogsRequest) (*jobpb.GetJobLogsResponse, error) {
	logs, err := h.dc.GetJobLogs(ctx, req.JobId)
	if err != nil {
		return nil, err
	}

	resp := &jobpb.GetJobLogsResponse{}
	for _, l := range logs {
		resp.Logs = append(resp.Logs, &jobpb.JobLog{
			Timestamp:    formatUTC(l.Timestamp),
			Status:       l.Status,
			ErrorMessage: l.ErrorMessage,
		})
	}

	return resp, nil
}

func (h *JobHandler) GetJobAdminLogs(ctx context.Context, req *jobpb.GetJobLogsAdminRequest) (*jobpb.GetJobLogsResponse, error) {
	logs, err := h.dc.GetJobAdminLogs(ctx, req.JobId)
	if err != nil {
		return nil, err
	}

	resp := &jobpb.GetJobLogsResponse{}
	for _, l := range logs {
		resp.Logs = append(resp.Logs, &jobpb.JobLog{
			Timestamp:    formatUTC(l.Timestamp),
			Status:       l.Status,
			ErrorMessage: l.ErrorMessage,
		})
	}

	return resp, nil
}

func (h *JobHandler) RetryJob(ctx context.Context, req *jobpb.RetryJobRequest) (*jobpb.RetryJobResponse, error) {

	id, err := h.dc.RetryJob(ctx, req.JobId)

	if err != nil {
		return nil, err
	}

	return &jobpb.RetryJobResponse{NewJobId: id}, nil
}

func (h *JobHandler) GetJobStats(ctx context.Context, req *jobpb.GetJobStatsRequest) (*jobpb.GetJobStatsResponse, error) {

	stats, err := h.dc.GetJobStats(ctx)

	if err != nil {
		return nil, err
	}

	return &jobpb.GetJobStatsResponse{
		TotalJobs:        stats.TotalJobs,
		TotalSuccessJobs: stats.TotalSuccessJobs,
		TotalFailedJobs:  stats.TotalFailedJobs,
	}, nil
}

func (h *JobHandler) GetJobMetrics(ctx context.Context, req *jobpb.GetJobMetricsRequest) (*jobpb.GetJobMetricsResponse, error) {
	data, err := h.dc.GetJobMetrics(ctx, timestampOrZero(req.From), timestampOrZero(req.To), req.Bucket)

	if err != nil {
		return nil, err
	}

	resp := &jobpb.GetJobMetricsResponse{}

	for _, item := range data {
		resp.Buckets = append(resp.Buckets, &jobpb.MetricBucket{
			Ts:        item.TS,
			Created:   item.Created,
			Completed: item.Completed,
			Failed:    item.Failed,
		})
	}

	return resp, nil
}

func mapJobs(jobs []domain.Job, total, limit, offset int) *jobpb.ListJobsResponse {
	resp := &jobpb.ListJobsResponse{}
	for _, j := range jobs {
		resp.Jobs = append(resp.Jobs, &jobpb.Job{
			JobId:       j.JobID,
			AppId:       j.AppID,
			Type:        j.Type,
			Status:      j.Status,
			Payload:     string(j.Payload),
			Retry:       int32(j.Retry),
			MaxRetry:    int32(j.MaxRetry),
			ScheduleAt:  formatUTC(j.ScheduledAt),
			CreatedAt:   formatUTC(j.CreatedAt),
			UpdatedAt:   formatUTC(j.UpdatedAt),
			ManualRetry: int32(j.ManualRetry),
		})
	}

	resp.Total = int32(total)
	resp.Limit = int32(limit)
	resp.Offset = int32(offset)

	return resp
}

func optionalTimestamp(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}

	value := ts.AsTime().UTC()
	return &value
}

func timestampOrZero(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}

	return ts.AsTime().UTC()
}

func formatUTC(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}

	return ts.UTC().Format(time.RFC3339)
}
