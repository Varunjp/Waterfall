package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"scheduler_service/internal/config"
	"scheduler_service/internal/consumer"
	"scheduler_service/internal/grpc/handler"
	pb "scheduler_service/internal/grpc/schedulerpb"
	"scheduler_service/internal/metrics"
	"scheduler_service/internal/repository"
	"scheduler_service/internal/usecase"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	cfg := config.Load()
	
	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal,1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutdown signal received")
		cancel()
	}()
	
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	jobStore := repository.NewRedisJobStore(rdb)
	workerRegistry := repository.NewRedisWorkerRegistry(rdb)
	quotaRepo := repository.NewRedisAdminQuotaRepo(rdb)

	eventQueue := repository.NewKafkaJobEventQueue(cfg.KafkaBroker,cfg.KafkaOutputTopic)
	logQueue := repository.NewKafkaJobLogQueue(cfg.KafkaBroker,cfg.KafkaLogTopic)

	schedulerMetrics := metrics.NewMetrics()

	schedulerUC := usecase.NewSchedulerUsecase(
		jobStore,
		workerRegistry,
		quotaRepo,
		eventQueue,
		schedulerMetrics,
		logQueue, 
	)

	jobReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBroker},
		Topic:   cfg.KafkaInputTopic,
		GroupID: "scheduler-service",
	})

	jobConsumer := consumer.NewJobCreatedConsumer(jobReader, jobStore)
	go func() {
		if err := jobConsumer.Start(ctx); err != nil {
			log.Println("job consumer stopped:", err)
		}
	}()

	go schedulerUC.StartStalledJobReaper(ctx)
	
	mux := http.NewServeMux()
	metricsServer := &http.Server{
		Addr: ":2112",
		Handler: mux,
	}

	go func() {
		log.Println("metrics listening on :2112")
		if err := metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Println("metrics server error:", err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(handler.GrpcUnaryInterceptor),
	)

	schedulerHandler := handler.NewSchedulerHandler(schedulerUC)
	pb.RegisterSchedulerServiceServer(grpcServer, schedulerHandler)
	
	reflection.Register(grpcServer)

	go func() {
		log.Println("scheduler gRPC listening on :" + cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Println("grpc server stopped:", err)
			cancel()
		}
	}()

	<-ctx.Done()

	log.Println("shutting down scheduler...")

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	grpcStopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcStopped)
	}()

	select {
	case <-grpcStopped:
		log.Println("grpc stopped gracefully")
	case <-shutdownCtx.Done():
		log.Println("grpc shutdown timed out, forcing stop")
		grpcServer.Stop()
	}
	_ = metricsServer.Shutdown(shutdownCtx)
	_ = rdb.Close()
	log.Println("scheduler stopped cleanly")
}