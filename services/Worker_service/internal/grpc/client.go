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
	conn,err := grpc.Dial(addr,grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	return &Client{
		api: pb.NewSchedulerClient(conn),
	}
}

func (c *Client) Heartbeat(ctx context.Context, jobID,appID,workerID string, progress int64) {
	c.api.Heartbeat(ctx,&pb.HeartbeatRequest{
		JobId: jobID,
		AppId: appID,
		WorkerId: workerID,
		Progress: progress,
		Timestamp: time.Now().Unix(),
	})
}

func (c *Client) ReportResult(ctx context.Context,jobID,appID,workerID string,success bool,errMsg string) {
	status := pb.JobResultStatus_JOB_RESULT_SUCCESS
	if !success {
		status = pb.JobResultStatus_JOB_RESULT_FAILED
	}

	c.api.ReportResult(ctx, &pb.JobResultRequest{
		JobId: jobID,
		AppId: appID,
		WorkerId: workerID,
		Status: status,
		ErrorMessage: errMsg,
		Timestamp: time.Now().Unix(),
	})
}