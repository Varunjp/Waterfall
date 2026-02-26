package grpcserver

import (
	"context"
	"net"
	"scheduler_service/internal/domain"
	schedulerpb "scheduler_service/internal/grpc/schedulerpb"
	"scheduler_service/internal/metrics"
	"scheduler_service/internal/producer"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	schedulerpb.UnimplementedSchedulerServer
	redis            *redis.Client
	producer         *producer.KafkaProducer
	metrics          *metrics.SchedulerMetrics
	jobResultProcess domain.JobResultUsecase
	log              *zap.Logger
}

func NewServer(redis *redis.Client, p *producer.KafkaProducer, m *metrics.SchedulerMetrics, j domain.JobResultUsecase, log *zap.Logger) *Server {
	return &Server{redis: redis, producer: p, metrics: m, jobResultProcess: j, log: log}
}

func (s *Server) Heartbeat(ctx context.Context, req *schedulerpb.HeartbeatRequest) (*schedulerpb.Ack, error) {

	key := "heartbeat:" + req.JobId

	err := s.redis.Set(
		ctx,
		key,
		req.Timestamp,
		30*time.Second,
	).Err()

	if err != nil {
		s.log.Warn("heartbeat failed", zap.Error(err))
	}

	if req.Progress == 0 {
		event := map[string]any{
			"job_id": req.JobId,
			"app_id": req.AppId,
			"status": "RUNNING",
			"retry":  0,
		}

		if err := s.producer.Publish(ctx, event); err != nil {
			s.log.Warn("failed to emit running update",
				zap.String("job_id", req.JobId),
				zap.Error(err),
			)
		}
	}

	return &schedulerpb.Ack{Ok: err == nil}, nil
}

func (s *Server) ReportResult(ctx context.Context, req *schedulerpb.JobResultRequest) (*schedulerpb.Ack, error) {
	jobID := req.JobId
	appID := req.AppId

	_ = s.redis.Del(ctx, "heartbeat:"+jobID)

	_ = s.redis.Decr(ctx, "concurrency:"+appID)

	status := "COMPLETED"
	if req.Status == schedulerpb.JobResultStatus_JOB_RESULT_FAILED {
		s.metrics.JobsFailed.Inc()
		status = "FAILED"
	} else {
		s.metrics.JobsSuccess.Inc()
	}
	s.metrics.RunningJobs.Dec()

	input := domain.JobResultInput{
		JobID:        jobID,
		AppID:        appID,
		Status:       status,
		Retry:        int(req.Retry),
		ErrorMessage: req.ErrorMessage,
	}

	if err := s.jobResultProcess.ProcessJobResult(ctx, input); err != nil {
		return &schedulerpb.Ack{Ok: false}, err
	}

	return &schedulerpb.Ack{Ok: true}, nil
}

func Run(ctx context.Context, addr string, redis *redis.Client, producer *producer.KafkaProducer, m *metrics.SchedulerMetrics, j domain.JobResultUsecase, log *zap.Logger) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	schedulerpb.RegisterSchedulerServer(grpcServer, NewServer(redis, producer, m, j, log))

	go func() {
		<-ctx.Done()
		log.Info("stopping grpc server")
		grpcServer.GracefulStop()
	}()

	log.Info("grpc server started", zap.String("addr", addr))
	return grpcServer.Serve(lis)
}
