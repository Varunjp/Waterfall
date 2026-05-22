package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"worker_service/internal/config"
	"worker_service/internal/executor"
	grpcclient "worker_service/internal/grpc"
	"worker_service/internal/logger"
	"worker_service/internal/worker"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("env not loaded")
	}

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// redisClient := redisClient.NewRedis(cfg.RedisAddr)
	// grpc := grpcclient.NewGrpcClient(cfg.SchedulerGRPC)

	// runner := worker.NewRunner(
	// 	redisClient,
	// 	grpc,
	// 	cfg.AppID,
	// 	cfg.WorkerID,
	// 	cfg.JobTypes,
	// 	cfg.MaxConcurrency,
	// 	cfg.HeartbeatInterval,
	// )
	// // delete
	// fmt.Println("worker started..")
	// runner.Run(ctx)

	log,err:= logger.NewLogger("worker")

	if err != nil {
		log.Fatal("error in logging",zap.Error(err))
	}

	client,err := grpcclient.NewGrpcClient(cfg.SchedulerGRPC,log)

	if err != nil {
		log.Fatal("grpc init failed",zap.Error(err))
	}

	runner := worker.NewRunner(client,executor.ExecuteJob,client.JobQueue,cfg.MaxConcurrency,log)

	runner.Start(ctx)

	errCh := make(chan error,1)

	go func() {
		errCh <- client.Start(ctx,cfg.AppID,cfg.WorkerID,cfg.JobTypes,cfg.MaxConcurrency)
	}()
	
	select {
	case <- ctx.Done():
		log.Info("shutdown requested")
	case err := <-errCh:
		log.Error("grpc client failed",zap.Error(err))
	}

	client.Close()
	runner.Wait()

	log.Info("worker shutdown complete")
}
