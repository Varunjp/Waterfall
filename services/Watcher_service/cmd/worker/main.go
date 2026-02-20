package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"watcher_service/internal/config"
	"watcher_service/internal/consumer"
	"watcher_service/internal/logger"
	"watcher_service/internal/producer"
	"watcher_service/internal/repository"
	"watcher_service/internal/usecase"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
)

// need to add redis for checking last poll time in case of failure

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system envs")
	}

	cfg := config.Load()
	validateConfig(cfg)

	logg,err := logger.Newlogger("watcher-service")
	defer logg.Sync() 
	
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	repo := repository.NewJobRepository(db)
	queueProducer := producer.NewKafkaProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	watchUC := usecase.NewWatchJobsUsecase(repo, queueProducer, logg)

	jobEventUC := usecase.NewConsumeJobUsecase(repo, logg)
	jobEventConsumer := consumer.NewKafkaConsumer(
		cfg.KafkaBroker,
		cfg.JobTopic,
		cfg.JobGroupID,
		jobEventUC,
		logg,
	)

	jobRunUC := usecase.NewUpdateJobStatusUsecase(repo)
	jobRunConsumer := consumer.NewJobRunConsumer(
		cfg.KafkaBroker,
		cfg.JobRunTopic,
		cfg.JobRunGroupID,
		jobRunUC,
		logg,
	)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := jobEventConsumer.Start(ctx); err != nil {
			logg.Error("job request consumer failed", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := jobRunConsumer.Start(ctx); err != nil {
			logg.Error("job run consumer failed", zap.Error(err))
		}
	}()

	ticker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer ticker.Stop()
		logg.Info("watcher poll loop started", zap.Int("interval_seconds", cfg.PollInterval))

		for {
			select {
			case <-ctx.Done():
				logg.Info("watcher poll loop shutting down")
				return
			case <-ticker.C:
				if err := watchUC.Run(ctx); err != nil {
					logg.Error("watcher tick failed", zap.Error(err))
				}
			}
		}
	}()

	logg.Info(
		"merged watcher+consumer service started",
		zap.String("watcher_topic", cfg.KafkaTopic),
		zap.String("job_topic", cfg.JobTopic),
		zap.String("job_run_topic", cfg.JobRunTopic),
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logg.Info("shutdown signal received")
	cancel()

	if err := jobEventConsumer.Close(); err != nil {
		logg.Error("job request consumer close failed", zap.Error(err))
	}

	if err := jobRunConsumer.Close(); err != nil {
		logg.Error("job run consumer close failed", zap.Error(err))
	}

	if err := queueProducer.Close(); err != nil {
		logg.Error("watcher producer close failed", zap.Error(err))
	}

	wg.Wait()
	logg.Info("merged service stopped cleanly")
}

func validateConfig(cfg *config.Config) {
	required := map[string]string{
		"DATABASE_URL":         cfg.DBURL,
		"KAFKA_BROKER":         firstBroker(cfg.KafkaBroker),
		"KAFKA_TOPIC":          cfg.KafkaTopic,
		"KAFKA_JOB_TOPIC":      cfg.JobTopic,
		"KAFKA_CONSUMER_GROUP": cfg.JobGroupID,
		"KAFKA_RUN_TOPIC":      cfg.JobRunTopic,
		"KAFKA_RUN_GROUP_ID":   cfg.JobRunGroupID,
	}

	for key, value := range required {
		if value == "" {
			log.Fatalf("missing required env var: %s", key)
		}
	}
}

func firstBroker(brokers []string) string {
	if len(brokers) == 0 {
		return ""
	}
	return brokers[0]
}
