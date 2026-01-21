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

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

	repo := postgresrepo.NewJobRepository(db)
	uc := usecase.NewConsumeJobUsecase(repo,logg)

	kafkaConsumer := consumer.NewKafkaConsumer(
		cfg.KafkaBroker,
		cfg.KafkaTopic,
		cfg.GroupID,
		uc,
		logg,
	)

	log.Println("Kafka consumer started")

	if err := kafkaConsumer.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}