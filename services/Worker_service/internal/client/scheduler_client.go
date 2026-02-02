package client

import (
	"context"
	"time"
	pb "worker_service/internal/grpc/schedulerpb"

	"google.golang.org/grpc"
)

type SchdeulerClient struct {
	pb.SchedulerServiceClient
	conn *grpc.ClientConn
}

func NewSchedulerClient(addr string)(*SchdeulerClient,error) {
	conn,err := grpc.Dial(addr,grpc.WithInsecure())
	if err != nil {
		return nil,err 
	}

	return &SchdeulerClient{
		SchedulerServiceClient: pb.NewSchedulerServiceClient(conn),
		conn: conn,
	},nil 
}

func (c *SchdeulerClient) Close(){
	_ = c.conn.Close()
}

func (c *SchdeulerClient) WithTimeout()(context.Context,context.CancelFunc) {
	return context.WithTimeout(context.Background(),30 * time.Second)
}