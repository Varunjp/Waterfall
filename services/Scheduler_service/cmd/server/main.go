package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"scheduler_service/internal/config"
	"scheduler_service/internal/consumer"
	grpcserver "scheduler_service/internal/grpc"
	"scheduler_service/internal/logger"
	"scheduler_service/internal/metrics"
	"scheduler_service/internal/producer"
	redisClient "scheduler_service/internal/redis"
	"scheduler_service/internal/repository"
	"scheduler_service/internal/scheduler"
	"scheduler_service/internal/usecase"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	
	rootCtx,stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := godotenv.Load(); err != nil {
		log.Println("env not loaded")
	}
	cfg := config.Load()
	log := logger.NewLogger()

	log.Info("scheduler starting")

	redisClient := redisClient.NewRedisClient(cfg.Redis.Addr)
	adminRepo := repository.NewAdminDb(cfg.AdminDB.DSN)

	kafkaConsumer := consumer.NewJobCreatedConsumer(cfg.Kafka.Brokers,cfg.Kafka.JobCreateTopic,cfg.Kafka.ConsumerGroup)
	kafkaProducer := producer.NewKafkaProducer([]string{cfg.Kafka.Brokers},cfg.Kafka.JobUpdateTopic)

	metrics := metrics.NewMetrics()

	assigner := usecase.NewAssigner(redisClient,adminRepo,metrics)
	stallMonitor := usecase.NewStallMonitor(redisClient,kafkaProducer)

	runner := scheduler.NewRunner(
		kafkaConsumer,assigner,kafkaProducer,redisClient.Client,log,
	)

	runtime := scheduler.NewRuntime(log)

	ctx := runtime.Start(
		rootCtx,
		runner.Run,
		stallMonitor.Run,

		func(c context.Context) {
			if err := grpcserver.Run(
				c,
				cfg.GRPC.Port,
				redisClient.Client,
				kafkaProducer,
				log,
			); err != nil {
				log.Error("grpc server failed",zap.Error(err))
			}
		},
	)

	<-ctx.Done()

	log.Info("shutdown signal received")
	runtime.Stop()
	log.Info("scheuler exited cleanly")
}