package main

import (
	"context"
	"database/sql"
	"log"
	"time"
	"watcher_service/internal/config"
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
	logg := logger.Newlogger()

	db, err := sql.Open("postgres",cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewJobRepository(db)
	producer := producer.NewKafkaProducer(cfg.KafkaBroker,cfg.KafkaTopic)
	uc := usecase.NewWatchJobsUsecase(repo,producer,logg)

	ticker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	log.Println("Watcher service started")

	for range ticker.C {
		if err := uc.Run(context.Background()); err != nil {
			logg.Error("watcher tick failed",zap.Error(err))
		}
	}
}