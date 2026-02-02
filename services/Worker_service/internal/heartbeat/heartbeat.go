package heartbeat

import (
	"context"
	"time"
	"worker_service/internal/client"
	"worker_service/internal/grpc/schedulerpb"
)

func Start(
	ctx context.Context,
	scheduler *client.SchdeulerClient,
	workerID string,
	interval time.Duration,
) {

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select{
		case <-ctx.Done():
			return 
		case <-ticker.C:
			_,_ = scheduler.WorkerHeartbeat(ctx,&schedulerpb.WorkerHeartbeatRequest{
				WorkerId: workerID,
			})
		}
	}
}