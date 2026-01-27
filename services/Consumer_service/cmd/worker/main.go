package main

import (
	"consumer_service/internal/config"
	"consumer_service/internal/consumer"
	"consumer_service/internal/logger"
	postgresrepo "consumer_service/internal/repository/postgres_repo"
	"consumer_service/internal/usecase"
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system envs")
	}

	cfg := config.Load()
	logg := logger.Newlogger()

	db,err := sql.Open("postgres",cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	repo := postgresrepo.NewJobRepository(db)
	uc := usecase.NewConsumeJobUsecase(repo,logg)
	jobrunUC := usecase.NewUpdateJobStatusUsecase(repo)

	kafkaConsumer := consumer.NewKafkaConsumer(
		cfg.KafkaBroker,
		cfg.KafkaTopic,
		cfg.GroupID,
		uc,
		logg,
	)

	jobRunConsumer := consumer.NewJobRunConsumer(
		cfg.KafkaBroker,
		cfg.KafkaRunTopic,
		cfg.RunGroupID,
		jobrunUC,
		logg,
	)

	ctx,cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	wg.Go(func() {
		kafkaConsumer.Start(ctx)
	})

	wg.Go(func() {
		jobRunConsumer.Start(ctx)
	})

	logg.Info("all Kafka consumer started")

	quit := make(chan os.Signal,1)
	signal.Notify(quit,os.Interrupt,syscall.SIGTERM)
	
	<-quit
	logg.Info("shutdown signal received")

	cancel()

	if err := jobRunConsumer.Close(); err != nil {
		logg.Error("consumer close failed",zap.Error(err))
	}

	wg.Wait()

	logg.Info("consumer service stopped cleanly")
}