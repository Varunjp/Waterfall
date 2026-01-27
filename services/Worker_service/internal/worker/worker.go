package worker

import (
	"context"
	"fmt"
	"time"
	"worker_service/internal/config"
	"worker_service/internal/executor"
	"worker_service/internal/heartbeat"
	"worker_service/internal/logger"
	"worker_service/internal/scheduler"
	pb "worker_service/proto/schedulerpb"

	"go.uber.org/zap"
)

type Worker struct {
	client *scheduler.Client
	cfg *config.Config
	log *zap.Logger
}

func NewWorker(client *scheduler.Client, cfg *config.Config) *Worker {
	return &Worker{
		client: client,
		cfg: cfg,
		log: logger.NewLog(),
	}
}

func (w *Worker) Run(ctx context.Context) {
	
	if err := w.Registerw(ctx); err != nil {
		w.log.Fatal("worker registration failed",zap.Error(err))
	}

	w.log.Info("worker started",
		zap.String("worker_id",w.cfg.WorkerID),
	)
	for {
		select {
		case <-ctx.Done():
			w.log.Info("worker shutting down")
			return 
		default:
			job,err := w.client.Poll(ctx,w.cfg.WorkerID,w.cfg.AppID,w.cfg.Capabilitis)
			//delete
			fmt.Println("job check in worker :",job)
			fmt.Println("err check in worker :",err)
			if err != nil || !job.Found {
				time.Sleep(2 * time.Second)
				continue 
			}

			w.handleJob(ctx,job)
		}
	}
}

func (w *Worker) handleJob(ctx context.Context, job *pb.PollJobResponse) {	
	log := w.log.With(
		zap.String("job_id",job.Job.JobId),
		zap.String("job_type",job.Job.JobType),
	)
	exec,ok := executor.Registry[job.Job.JobType]
	if !ok {
		log.Error("unsupported job type")
		w.fail(ctx,job.Job.JobId, "unsupported job type")
		return 
	}

	jobCtx,cancel := context.WithCancel(ctx)
	defer cancel()

	heartbeat.Start(jobCtx,w.client,job.Job.JobId,w.cfg.WorkerID,w.cfg.HeartbeatSec)

	err := executor.SafeExecute(jobCtx,func(ctx context.Context)error{
		return exec.Execute(jobCtx,job.Job.Payload)
	},log)
	
	if err != nil {
		log.Error("job failed",zap.Error(err))
		w.fail(ctx,job.Job.JobId,err.Error())
		return 
	}

	log.Info("job completed")
	w.complete(ctx,job.Job.JobId)
}

func (w *Worker) complete(ctx context.Context, jobID string) {
	w.client.CompleteJob(ctx,&pb.CompleteJobRequest{
		JobId: jobID,
		WorkerId: w.cfg.WorkerID,
		AppId: w.cfg.AppID,
		Result: pb.JobResult_JOB_RESULT_SUCCESS,
	})
}

func (w *Worker) fail(ctx context.Context,jobID,reason string) {
	w.client.JobHeartbeat(ctx,&pb.JobHeartbeatRequest{
		JobId: jobID,
		WorkerId: w.cfg.WorkerID,
		Message: reason,
	})
}