package grpcclient

import (
	"context"
	"time"
	pb "worker_service/internal/grpc/schedulerpb"

	"google.golang.org/grpc"
)

type Client struct {
	api pb.SchedulerClient
}

func NewGrpcClient(addr string) *Client {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	return &Client{
		api: pb.NewSchedulerClient(conn),
	}
}

func (c *Client) Heartbeat(ctx context.Context, jobID, appID, workerID string, progress int64) {
	if _, err := c.api.Heartbeat(ctx, &pb.HeartbeatRequest{
		JobId:     jobID,
		AppId:     appID,
		WorkerId:  workerID,
		Progress:  progress,
		Timestamp: time.Now().Unix(),
	}); err != nil {
		return
	}
}

func (c *Client) ReportResult(ctx context.Context, jobID, appID, workerID string, success bool, errMsg string, retry int, manual_retry int) {
	status := pb.JobResultStatus_JOB_RESULT_SUCCESS
	if !success {
		status = pb.JobResultStatus_JOB_RESULT_FAILED
	}

	if _, err := c.api.ReportResult(ctx, &pb.JobResultRequest{
		JobId:        jobID,
		AppId:        appID,
		WorkerId:     workerID,
		Status:       status,
		ErrorMessage: errMsg,
		Retry:        int32(retry),
		ManualRetry:  int32(manual_retry),
		Timestamp:    time.Now().Unix(),
	}); err != nil {
		return
	}
}

func (c *Client) RegisterWorker(ctx context.Context, appID, workerID string, jobTypes []string, maxConcurrency int) {
	_, _ = c.api.RegisterWorker(ctx, &pb.RegisterWorkerRequest{
		AppId:          appID,
		WorkerId:       workerID,
		JobTypes:       jobTypes,
		MaxConcurrency: int32(maxConcurrency),
		Timestamp:      time.Now().Unix(),
	})
}

func (c *Client) WorkerHeartbeat(ctx context.Context, appID, workerID string, jobTypes []string, activeJobs int, maxConcurrency int) {
	_, _ = c.api.WorkerHeartbeat(ctx, &pb.WorkerHeartbeatRequest{
		AppId:          appID,
		WorkerId:       workerID,
		JobTypes:       jobTypes,
		ActiveJobs:     int32(activeJobs),
		MaxConcurrency: int32(maxConcurrency),
		Timestamp:      time.Now().Unix(),
	})
}

func (c *Client) UnregisterWorker(ctx context.Context, appID, workerID string) {
	_, _ = c.api.UnregisterWorker(ctx, &pb.UnregisterWorkerRequest{
		AppId:     appID,
		WorkerId:  workerID,
		Timestamp: time.Now().Unix(),
	})
}
