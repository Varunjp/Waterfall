package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"worker_service/internal/config"
	"worker_service/internal/logger"
	"worker_service/internal/scheduler"
	"worker_service/internal/worker"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("env not loaded")
	}
	logger.Init()
	cfg := config.Load()

	client,err := scheduler.NewScheduler(cfg.SchedulerAddr)
	if err != nil {
		panic(err)
	}

	w := worker.NewWorker(client,cfg)

	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Run(ctx)

	sig := make(chan os.Signal,1)
	signal.Notify(sig,syscall.SIGINT,syscall.SIGTERM)
	<-sig 
	cancel()
}