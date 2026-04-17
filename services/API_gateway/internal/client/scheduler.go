package client

import (
	"context"

	schedulerpb "api_gateway/internal/proto/schedulerpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type SchedulerClient struct {
	api schedulerpb.SchedulerClient
}

func NewScheduler(endpoint string) (*SchedulerClient, error) {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &SchedulerClient{
		api: schedulerpb.NewSchedulerClient(conn),
	}, nil
}

func (c *SchedulerClient) GetTenantRuntime(
	ctx context.Context,
	authorization string,
	appID string,
) (*schedulerpb.GetTenantRuntimeResponse, error) {
	if authorization != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authorization)
	}

	return c.api.GetTenantRuntime(ctx, &schedulerpb.GetTenantRuntimeRequest{
		AppId: appID,
	})
}
