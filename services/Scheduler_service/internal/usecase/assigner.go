package usecase

import (
	"context"
	"log"
	"scheduler_service/internal/domain"
	grpcserver "scheduler_service/internal/grpc"
	"scheduler_service/internal/grpc/schedulerpb"
	"scheduler_service/internal/metrics"
	"scheduler_service/internal/monitoring"
	"scheduler_service/internal/producer"
	redisClient "scheduler_service/internal/redis"
	"time"
)

type Assigner struct {
	redis    *redisClient.Client
	metrics  *metrics.SchedulerMetrics
	producer *producer.KafkaProducer
	runtime  *monitoring.Store

	workerMg *grpcserver.WorkerManger
	dispatch grpcserver.GrpcDispatch
	jobuse   *jobUsecase
}

func NewAssigner(
	r *redisClient.Client,
	m *metrics.SchedulerMetrics,
	p *producer.KafkaProducer,
	store *monitoring.Store,
	w *grpcserver.WorkerManger,
	d grpcserver.GrpcDispatch,
	jobu *jobUsecase,
) *Assigner {
	return &Assigner{
		redis:    r,
		metrics:  m,
		producer: p,
		runtime:  store,
		workerMg: w,
		dispatch: d,
		jobuse:   jobu,
	}
}

func (a *Assigner) Assign(ctx context.Context, job domain.Job) error {
	// stream := fmt.Sprintf("stream:jobs:%s:%s", job.AppID, job.Type)
	// group := fmt.Sprintf("workers:%s:%s", job.AppID, job.Type)

	// if err := redisClient.EnsureGroup(ctx, a.redis.Client, stream, group); err != nil {
	// 	return err
	// }

	// key := fmt.Sprintf("concurrency:%s", job.AppID)

	// _, err := a.redis.EvalSha(
	// 	ctx,
	// 	a.redis.AssignSHA,
	// 	[]string{key, stream},
	// 	job.JobID,
	// 	job.Payload,
	// 	job.Retry,
	// 	job.ManualRetry,
	// ).Result()

	worker := a.workerMg.FindAvailableWorker(job.AppID, job.Type)

	if worker == nil {
		jobRes := domain.JobResultInput{
			JobID:        job.JobID,
			AppID:        job.AppID,
			Retry:        job.Retry,
			Status:       string(domain.JobFailed),
			ErrorMessage: "No worker available",
		}
		a.metrics.JobsFailed.Inc()
		return a.jobuse.ProcessJobResult(ctx, jobRes)
	} else {

		msg := &schedulerpb.SchedulerMessage{
			Payload: &schedulerpb.SchedulerMessage_Job{
				Job: &schedulerpb.JobAssignment{
					JobId:      job.JobID,
					AppId:      job.AppID,
					JobType:    job.Type,
					Payload:    job.Payload,
					RetryCount: int32(job.Retry),
					MaxRetries: int32(job.MaxRetries),
				},
			},
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-worker.Ctx.Done():
			return domain.ErrWorkerDisconnected
		case worker.SendQueue <- msg:
			worker.IncrementJobs()
			a.metrics.RunningJobs.Inc()
			a.metrics.JobsAssigned.Inc()
			if a.runtime != nil {
				err := a.runtime.RecordQueuedJob(ctx, job, time.Now().UTC())
				if err != nil {
					log.Println("failed to record queued job",err)
				}
			}
			return nil
		case <-time.After(3 * time.Second):
			return domain.ErrWorkerQueueFull
		}
	}
}
