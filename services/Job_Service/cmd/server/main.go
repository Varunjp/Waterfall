package main

import (
	"context"
	"job_service/internal/config"
	"job_service/internal/idempotency"
	"job_service/internal/infrastructure/kafka"
	"job_service/internal/infrastructure/redis"
	"job_service/internal/interceptor"
	jobpb "job_service/internal/proto"
	"job_service/internal/server"
	"job_service/internal/telemetry"
	"job_service/internal/usecase"
	"log"
	"net"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file foudn, using system envs")
	}
	
	cfg := config.Load()

	log,_ := zap.NewProduction()
	defer log.Sync()

	ctx := context.Background()
	shutdown,err := telemetry.Init(ctx,cfg.ServiceName,cfg.OtelEndpoint)
	if err != nil {
		log.Fatal("otel init failed",zap.Error(err))
	}
	defer shutdown(ctx)

	rdb := redis.Newredis(cfg.RedisAddr)
	idem := idempotency.New(rdb)
	producer := kafka.NewProducer(cfg.KafkaBrokers)

	uc := usecase.NewJobUsecase(producer,idem)

	lis,err := net.Listen("tcp",":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("listen failed",zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.UnaryLogger(log)),
	)

	jobpb.RegisterJobServiceServer(grpcServer,server.New(uc))

	log.Info("job service started",zap.String("port",cfg.GRPCPort))
	grpcServer.Serve(lis)
}