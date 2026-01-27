package heartbeat

import (
	"context"
	"time"
	pb "worker_service/proto/schedulerpb"
)
func Start(ctx context.Context, client pb.SchedulerServiceClient, jobID, workerID string, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				client.JobHeartbeat(ctx,&pb.JobHeartbeatRequest{
					JobId: jobID,
					WorkerId: workerID,
					Message: "running",
				})
			case <-ctx.Done():
				return 
			}
		}
	}()
}