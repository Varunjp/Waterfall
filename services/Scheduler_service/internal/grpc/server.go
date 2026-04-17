package grpcserver

import (
	"context"
	"net"
	"scheduler_service/internal/domain"
	schedulerpb "scheduler_service/internal/grpc/schedulerpb"
	"scheduler_service/internal/metrics"
	"scheduler_service/internal/middleware"
	"scheduler_service/internal/monitoring"
	"scheduler_service/internal/producer"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	schedulerpb.UnimplementedSchedulerServer
	redis            *redis.Client
	producer         *producer.KafkaProducer
	metrics          *metrics.SchedulerMetrics
	jobResultProcess domain.JobResultUsecase
	log              *zap.Logger
	runtime          *monitoring.Store
}

func NewServer(
	redis *redis.Client,
	p *producer.KafkaProducer,
	m *metrics.SchedulerMetrics,
	j domain.JobResultUsecase,
	log *zap.Logger,
	store *monitoring.Store,
) *Server {
	return &Server{
		redis:            redis,
		producer:         p,
		metrics:          m,
		jobResultProcess: j,
		log:              log,
		runtime:          store,
	}
}

func (s *Server) Heartbeat(ctx context.Context, req *schedulerpb.HeartbeatRequest) (*schedulerpb.Ack, error) {
	now := requestTime(req.Timestamp)
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

	if s.runtime != nil {
		if err := s.runtime.RecordWorkerSeen(ctx, req.AppId, req.WorkerId, now); err != nil {
			s.log.Warn("failed to refresh worker activity",
				zap.String("worker_id", req.WorkerId),
				zap.String("app_id", req.AppId),
				zap.Error(err),
			)
		}
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

		if s.runtime != nil {
			if err := s.runtime.RecordJobStarted(ctx, req.AppId, req.WorkerId, req.JobId, now); err != nil {
				s.log.Warn("failed to record job start",
					zap.String("job_id", req.JobId),
					zap.String("worker_id", req.WorkerId),
					zap.Error(err),
				)
			}
		}
	}

	return &schedulerpb.Ack{Ok: err == nil}, nil
}

func (s *Server) ReportResult(ctx context.Context, req *schedulerpb.JobResultRequest) (*schedulerpb.Ack, error) {
	jobID := req.JobId
	appID := req.AppId
	now := requestTime(req.Timestamp)

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

	if s.runtime != nil {
		if err := s.runtime.RecordJobFinished(ctx, appID, req.WorkerId, jobID, now); err != nil {
			s.log.Warn("failed to record job completion",
				zap.String("job_id", jobID),
				zap.String("worker_id", req.WorkerId),
				zap.Error(err),
			)
		}
	}

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

func (s *Server) RegisterWorker(ctx context.Context, req *schedulerpb.RegisterWorkerRequest) (*schedulerpb.Ack, error) {
	if req.AppId == "" || req.WorkerId == "" {
		return &schedulerpb.Ack{Ok: false}, status.Error(codes.InvalidArgument, "app_id and worker_id are required")
	}

	if s.runtime == nil {
		return &schedulerpb.Ack{Ok: false}, status.Error(codes.FailedPrecondition, "runtime store not configured")
	}

	if err := s.runtime.RecordWorkerRegistration(
		ctx,
		req.AppId,
		req.WorkerId,
		req.JobTypes,
		int(req.MaxConcurrency),
		requestTime(req.Timestamp),
	); err != nil {
		return &schedulerpb.Ack{Ok: false}, err
	}

	return &schedulerpb.Ack{Ok: true}, nil
}

func (s *Server) WorkerHeartbeat(ctx context.Context, req *schedulerpb.WorkerHeartbeatRequest) (*schedulerpb.Ack, error) {
	if req.AppId == "" || req.WorkerId == "" {
		return &schedulerpb.Ack{Ok: false}, status.Error(codes.InvalidArgument, "app_id and worker_id are required")
	}

	if s.runtime == nil {
		return &schedulerpb.Ack{Ok: false}, status.Error(codes.FailedPrecondition, "runtime store not configured")
	}

	if err := s.runtime.RecordWorkerHeartbeat(
		ctx,
		req.AppId,
		req.WorkerId,
		req.JobTypes,
		int(req.MaxConcurrency),
		int(req.ActiveJobs),
		requestTime(req.Timestamp),
	); err != nil {
		return &schedulerpb.Ack{Ok: false}, err
	}

	return &schedulerpb.Ack{Ok: true}, nil
}

func (s *Server) UnregisterWorker(ctx context.Context, req *schedulerpb.UnregisterWorkerRequest) (*schedulerpb.Ack, error) {
	if req.AppId == "" || req.WorkerId == "" {
		return &schedulerpb.Ack{Ok: false}, status.Error(codes.InvalidArgument, "app_id and worker_id are required")
	}

	if s.runtime == nil {
		return &schedulerpb.Ack{Ok: false}, status.Error(codes.FailedPrecondition, "runtime store not configured")
	}

	if err := s.runtime.MarkWorkerOffline(ctx, req.AppId, req.WorkerId, requestTime(req.Timestamp)); err != nil {
		return &schedulerpb.Ack{Ok: false}, err
	}

	return &schedulerpb.Ack{Ok: true}, nil
}

func (s *Server) GetTenantRuntime(ctx context.Context, req *schedulerpb.GetTenantRuntimeRequest) (*schedulerpb.GetTenantRuntimeResponse, error) {
	if s.runtime == nil {
		return nil, status.Error(codes.FailedPrecondition, "runtime store not configured")
	}

	targetAppID, err := resolveTargetAppID(ctx, req)
	if err != nil {
		return nil, err
	}

	snapshot, err := s.runtime.SnapshotTenant(ctx, targetAppID, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	resp := &schedulerpb.GetTenantRuntimeResponse{
		AppId:            snapshot.AppID,
		GeneratedAt:      snapshot.GeneratedAt.Unix(),
		TotalReadyJobs:   snapshot.TotalReadyJobs,
		TotalRunningJobs: snapshot.TotalRunningJobs,
		TotalWorkers:     snapshot.TotalWorkers,
		OnlineWorkers:    snapshot.OnlineWorkers,
		BusyWorkers:      snapshot.BusyWorkers,
	}

	for _, queue := range snapshot.Queues {
		resp.Queues = append(resp.Queues, &schedulerpb.QueueRuntime{
			JobType:               queue.JobType,
			ReadyJobs:             queue.ReadyJobs,
			RunningJobs:           queue.RunningJobs,
			RegisteredWorkers:     queue.RegisteredWorkers,
			BusyWorkers:           queue.BusyWorkers,
			OldestReadyAgeSeconds: queue.OldestReadyAgeSeconds,
		})
	}

	for _, worker := range snapshot.Workers {
		resp.Workers = append(resp.Workers, &schedulerpb.WorkerRuntime{
			WorkerId:       worker.WorkerID,
			JobTypes:       worker.JobTypes,
			ActiveJobs:     int32(worker.ActiveJobs),
			MaxConcurrency: int32(worker.MaxConcurrency),
			LastSeen:       worker.LastSeen.Unix(),
			Status:         mapWorkerStatus(worker.Status),
		})
	}

	return resp, nil
}

func Run(
	ctx context.Context,
	addr string,
	redis *redis.Client,
	producer *producer.KafkaProducer,
	m *metrics.SchedulerMetrics,
	j domain.JobResultUsecase,
	log *zap.Logger,
	store *monitoring.Store,
	jwtSecret string,
) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(middleware.APIKeyInterceptor(jwtSecret)),
	)
	schedulerpb.RegisterSchedulerServer(grpcServer, NewServer(redis, producer, m, j, log, store))

	go func() {
		<-ctx.Done()
		log.Info("stopping grpc server")
		grpcServer.GracefulStop()
	}()

	log.Info("grpc server started", zap.String("addr", addr))
	return grpcServer.Serve(lis)
}

func resolveTargetAppID(ctx context.Context, req *schedulerpb.GetTenantRuntimeRequest) (string, error) {
	role, err := middleware.RoleFromContext(ctx)
	if err != nil {
		return "", status.Error(codes.PermissionDenied, "missing role")
	}

	requestedAppID := req.GetAppId()
	claimedAppID, _ := middleware.AppIDFromContext(ctx)

	if role == "platform_admin" {
		if requestedAppID == "" {
			return "", status.Error(codes.InvalidArgument, "app_id is required for platform_admin")
		}
		return requestedAppID, nil
	}

	if claimedAppID == "" {
		return "", status.Error(codes.PermissionDenied, "missing tenant app_id")
	}

	if requestedAppID != "" && requestedAppID != claimedAppID {
		return "", status.Error(codes.PermissionDenied, "cannot access another tenant runtime")
	}

	return claimedAppID, nil
}

func mapWorkerStatus(status domain.WorkerRuntimeStatus) schedulerpb.WorkerStatus {
	switch status {
	case domain.WorkerStatusOnline:
		return schedulerpb.WorkerStatus_WORKER_STATUS_ONLINE
	case domain.WorkerStatusBusy:
		return schedulerpb.WorkerStatus_WORKER_STATUS_BUSY
	case domain.WorkerStatusStale:
		return schedulerpb.WorkerStatus_WORKER_STATUS_STALE
	case domain.WorkerStatusOffline:
		return schedulerpb.WorkerStatus_WORKER_STATUS_OFFLINE
	default:
		return schedulerpb.WorkerStatus_WORKER_STATUS_UNKNOWN
	}
}

func requestTime(ts int64) time.Time {
	if ts <= 0 {
		return time.Now().UTC()
	}
	return time.Unix(ts, 0).UTC()
}
