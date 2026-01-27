package scheduler

import (
	"context"
	pb "worker_service/proto/schedulerpb"

	"google.golang.org/grpc"
)

type Client struct {
	pb.SchedulerServiceClient
}

func NewScheduler(addr string)(*Client,error) {
	conn, err := grpc.Dial(addr,grpc.WithInsecure())
	if err != nil {
		return nil,err 
	}

	return &Client{
		SchedulerServiceClient: pb.NewSchedulerServiceClient(conn),
	},nil 
}

func (c *Client) Poll(ctx context.Context, workerID, appID string, caps []string) (*pb.PollJobResponse,error) {
	return c.PollJob(ctx,&pb.PollJobRequest{
		WorkerId: workerID,
		AppId: appID,
		Capabilities: caps,
	})
}

func (c *Client) Registerworker(ctx context.Context,workerID,appID string,capabilities []string, maxJobs int) error {
	
	_,err := c.RegisterWorker(ctx,&pb.RegisterWorkerRequest{
		WorkerId: workerID,
		AppId: appID,
		Capabilities: capabilities,
		MaxConcurrency: int32(maxJobs),
		HeartbeatIntervalSec: 20,
	})

	return err 
}