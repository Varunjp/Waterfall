package worker

import (
	"context"
	"time"

	"go.uber.org/zap"
)

func (w *Worker) Registerw(ctx context.Context)error {
	w.log.Info("registering worker",
		zap.String("worker_id",w.cfg.WorkerID),
		zap.String("app_id",w.cfg.AppID),
		zap.Strings("capabilites",w.cfg.Capabilitis),
		zap.Int("max_concurrent_jobs",w.cfg.MaxConcurrentJobs),
	)

	ctx,cancel := context.WithTimeout(ctx, 5 *time.Second)
	defer cancel()

	return  w.client.Registerworker(
		ctx,
		w.cfg.WorkerID,
		w.cfg.AppID,
		w.cfg.Capabilitis,
		w.cfg.MaxConcurrentJobs,
	)
}