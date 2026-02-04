package grpcserver

import (
	"context"
	"net"
	schedulerpb "scheduler_service/internal/grpc/schedulerpb"
	"scheduler_service/internal/producer"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	schedulerpb.UnimplementedSchedulerServer
	redis *redis.Client
	producer *producer.KafkaProducer
	log *zap.Logger

}

func NewServer(redis *redis.Client,p *producer.KafkaProducer , log *zap.Logger) *Server {
	return &Server{redis: redis,producer: p,log: log}
}

func (s *Server) Heartbeat(ctx context.Context, req *schedulerpb.HeartbeatRequest)(*schedulerpb.Ack,error) {

	key := "heartbeat:"+req.JobId

	err := s.redis.Set(
		ctx,
		key,
		req.Timestamp,
		30*time.Second,
	).Err()

	if err != nil {
		s.log.Warn("heartbeat failed",zap.Error(err))
	}

	return &schedulerpb.Ack{Ok: err==nil},nil 
}

func (s *Server) ReportResult(ctx context.Context,req *schedulerpb.JobResultRequest)(*schedulerpb.Ack,error) {
	jobID := req.JobId
	appID := req.AppId

	_ = s.redis.Del(ctx,"heartbeat:"+jobID)

	_ = s.redis.Decr(ctx,"concurrency:"+appID)

	status := "COMPLETED"
	if req.Status == schedulerpb.JobResultStatus_JOB_RESULT_FAILED {
		status = "FAILED"
	}

	event := map[string]any{
		"job_id": jobID,
		"app_id": appID,
		"status": status,
		"error": req.ErrorMessage,
	}

	if err := s.producer.Publish(ctx,event); err != nil {
		s.log.Error("failed to publish job result",zap.Error(err))
		return &schedulerpb.Ack{Ok: false},err 
	}

	s.log.Info("job result processed",
		zap.String("job_id",jobID),
		zap.String("status",status),
	)

	return &schedulerpb.Ack{Ok: true},nil 
}

func Run(ctx context.Context,addr string,redis *redis.Client,producer *producer.KafkaProducer,log *zap.Logger) error {
	lis,err := net.Listen("tcp",addr)
	if err != nil {
		return err 
	}

	grpcServer := grpc.NewServer()
	schedulerpb.RegisterSchedulerServer(grpcServer,NewServer(redis,producer,log))

	go func() {
		<-ctx.Done()
		log.Info("stopping grpc server")
		grpcServer.GracefulStop()
	}()

	log.Info("grpc server started",zap.String("addr",addr))
	return grpcServer.Serve(lis)
}