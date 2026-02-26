package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"scheduler_service/internal/consumer"
	"scheduler_service/internal/domain"
	"scheduler_service/internal/producer"
	"scheduler_service/internal/usecase"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrConcurrencyLimit = errors.New("concurrency limit reached")
)

type Runner struct {
	kafak *consumer.JobCreatedConsumer
	assigner *usecase.Assigner
	producer *producer.KafkaProducer
	redis *redis.Client
	log  *zap.Logger
}

func NewRunner(c *consumer.JobCreatedConsumer,a *usecase.Assigner,p *producer.KafkaProducer,r *redis.Client,l *zap.Logger) *Runner {
	return &Runner{
		kafak: c,
		assigner: a,
		producer: p,
		redis: r,
		log: l,
	}
}

func (r *Runner) Run(ctx context.Context) {
	r.log.Info("scheduler runner started")
	for {
		select {
		case <-ctx.Done():
			r.log.Info("scheduler loop stopped")
			return 
		default:
			msg,err := r.kafak.Read(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return 
				}
				r.log.Error("kafka read failed",zap.Error(err))
				continue
			}
			r.handleMessage(ctx,msg)
		}
	}
}

func (r *Runner) handleMessage(ctx context.Context,raw []byte) {
	var job domain.Job
	if err := json.Unmarshal(raw,&job); err != nil {
		r.log.Warn("invalid job payload",zap.Error(err))
		return
	}

	err := r.assigner.Assign(ctx,job)
	
	if err != nil {
		r.handleAssignFailure(ctx,job,err)
		return 
	}

	r.presistRunningJob(ctx,job)

	//r.emitJobUpdate(ctx,job,"RUNNING","")
}

func (r *Runner) handleAssignFailure(ctx context.Context,job domain.Job,err error) {

	if errors.Is(err,ErrConcurrencyLimit) {
		r.log.Debug("concurrency limit hit",
			zap.String("job_id",job.JobID),
			zap.String("app_id",job.AppID),
		)
		return 
	}

	r.log.Error("job assignment failed", 
		zap.String("job_id",job.JobID),
		zap.Error(err),
	)

	r.emitJobUpdate(ctx,job,"FAILED",err.Error())
}

func (r *Runner) presistRunningJob(ctx context.Context, job domain.Job) {
	data, _ := json.Marshal(job)

	err := r.redis.Set(
		ctx,
		"running:"+job.JobID,
		data,
		0,
	).Err()

	if err != nil {
		r.log.Warn("failed to persist running job",
			zap.String("job_id",job.JobID),
			zap.Error(err),
		)
	}
}

func (r *Runner) emitJobUpdate(
	ctx context.Context,
	job domain.Job,
	status string,
	errMsg string,
) {
	event := map[string]any{
		"job_id": job.JobID,
		"app_id": job.AppID,
		"status": status,
		"retries": job.Retry,
		"error": errMsg,
	}

	if err := r.producer.Publish(ctx,event); err != nil {
		r.log.Warn("failed to emit job update", 
			zap.String("job_id",job.JobID),
			zap.String("status",status),
			zap.Error(err),
		)
	}
}