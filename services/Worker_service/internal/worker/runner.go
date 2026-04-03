package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"worker_service/internal/domain"
	"worker_service/internal/executor"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

type Runner struct {
	redis             *redis.Client
	grpc              GRPC
	appID             string
	workerID          string
	jobTypes          []string
	maxConcurrency    int
	heartbeatInterval time.Duration
	activeJobs        int32
}

type GRPC interface {
	Heartbeat(ctx context.Context, jobID, appID, workerID string, progress int64)
	ReportResult(ctx context.Context, jobID, appID, workerID string, success bool, errMsg string, retry int, manual_retry int)
	RegisterWorker(ctx context.Context, appID, workerID string, jobTypes []string, maxConcurrency int)
	WorkerHeartbeat(ctx context.Context, appID, workerID string, jobTypes []string, activeJobs int, maxConcurrency int)
	UnregisterWorker(ctx context.Context, appID, workerID string)
}

func NewRunner(
	redis *redis.Client,
	grpc GRPC,
	appID, workerID string,
	jobTypes []string,
	maxConcurrency int,
	heartbeatInterval time.Duration,
) *Runner {
	return &Runner{
		redis:             redis,
		grpc:              grpc,
		appID:             appID,
		workerID:          workerID,
		jobTypes:          jobTypes,
		maxConcurrency:    maxConcurrency,
		heartbeatInterval: heartbeatInterval,
	}
}

func (r *Runner) Run(ctx context.Context) {
	r.grpc.RegisterWorker(ctx, r.appID, r.workerID, r.jobTypes, r.maxConcurrency)
	go r.workerHeartbeatLoop(ctx)

	for _, jobType := range r.jobTypes {
		go r.consume(ctx, jobType)
	}
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r.grpc.UnregisterWorker(shutdownCtx, r.appID, r.workerID)
}

func (r *Runner) consume(ctx context.Context, jobType string) {
	stream := "stream:jobs:" + r.appID + ":" + jobType
	group := "workers:" + r.appID + ":" + jobType
	err := r.redis.XGroupCreateMkStream(stream, group, "0").Err()

	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		panic(err)
	}

	r.read(ctx, stream, group, "0")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			r.read(ctx, stream, group, ">")
		}
	}
}

func (r *Runner) read(ctx context.Context, stream, group, id string) {
	res, err := r.redis.XReadGroup(&redis.XReadGroupArgs{
		Group:    group,
		Consumer: r.workerID,
		Streams:  []string{stream, id},
		Count:    1,
		Block:    5 * time.Second,
	}).Result()

	if err != nil || len(res) == 0 {
		return
	}

	for _, msg := range res[0].Messages {
		job, err := parseJob(msg)
		if err != nil {
			log.Println("failed to parse job", zap.Error(err))
			//r.redis.XAck(stream, group, msg.ID)
			return
		}

		r.handleJob(ctx, stream, group, msg.ID, job)
	}
}

func (r *Runner) handleJob(ctx context.Context, stream, group, msgID string, job *domain.Job) {
	hbCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	atomic.AddInt32(&r.activeJobs, 1)
	defer atomic.AddInt32(&r.activeJobs, -1)

	r.grpc.Heartbeat(hbCtx, job.JobID, r.appID, r.workerID, 0)

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-hbCtx.Done():
				return
			case <-ticker.C:
				r.grpc.Heartbeat(hbCtx, job.JobID, r.appID, r.workerID, 50)
			}
		}
	}()

	err := executor.Execute(ctx, job.Payload)

	if err != nil {
		r.grpc.ReportResult(ctx, job.JobID, r.appID, r.workerID, false, err.Error(), job.Retry, job.ManualRetry)
	} else {
		r.grpc.ReportResult(ctx, job.JobID, r.appID, r.workerID, true, "", job.Retry, job.ManualRetry)
	}

	r.redis.XAck(stream, group, msgID)
}

func (r *Runner) workerHeartbeatLoop(ctx context.Context) {
	interval := r.heartbeatInterval
	if interval <= 0 {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	r.grpc.WorkerHeartbeat(ctx, r.appID, r.workerID, r.jobTypes, int(atomic.LoadInt32(&r.activeJobs)), r.maxConcurrency)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.grpc.WorkerHeartbeat(ctx, r.appID, r.workerID, r.jobTypes, int(atomic.LoadInt32(&r.activeJobs)), r.maxConcurrency)
		}
	}
}

func parseJob(msg redis.XMessage) (*domain.Job, error) {
	rawJobID, ok := msg.Values["job_id"]
	if !ok {
		return nil, errors.New("missing job_id in stream message")
	}

	jobID, ok := rawJobID.(string)
	if !ok {
		return nil, fmt.Errorf("job_id field is not string: %T", rawJobID)
	}
	rawPayload, ok := msg.Values["payload"]
	if !ok {
		return nil, errors.New("missing payload in stream message")
	}
	payloadStr, ok := rawPayload.(string)
	if !ok {
		return nil, fmt.Errorf("payload is not string: %T", rawPayload)
	}
	rawAttempt, ok := msg.Values["attempt"]
	if !ok {
		return nil, errors.New("missing attempt in stream message")
	}
	attempt, err := toInt(rawAttempt)
	if err != nil {
		return nil, fmt.Errorf("invalid attempt: %w", err)
	}
	job := &domain.Job{
		JobID:   jobID,
		Payload: payloadStr,
		Retry:   attempt,
	}

	return job, nil
}

func toInt(v interface{}) (int, error) {
	switch t := v.(type) {
	case int:
		return t, nil
	case int64:
		return int(t), nil
	case string:
		return strconv.Atoi(t)
	case []byte:
		return strconv.Atoi(string(t))
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}
