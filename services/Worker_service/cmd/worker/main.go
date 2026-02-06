package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"worker_service/internal/config"
	grpcclient "worker_service/internal/grpc"
	redisClient "worker_service/internal/redis"
	"worker_service/internal/worker"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("env not loaded")
	}

	cfg := config.Load()

	ctx,stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	redisClient := redisClient.NewRedis(cfg.RedisAddr)
	grpc := grpcclient.NewGrpcClient(cfg.SchedulerGRPC)

	runner := worker.NewRunner(
		redisClient,
		grpc,
		cfg.AppID,
		cfg.WorkerID,
		cfg.JobTypes,
	)
	// delete
	fmt.Println("worker started..")
	runner.Run(ctx)
}