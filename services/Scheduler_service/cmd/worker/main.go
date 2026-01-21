package main

import (
	"context"
	"log"
	"scheduler_service/internal/config"
	"scheduler_service/internal/consumer"
	"scheduler_service/internal/logger"
	"scheduler_service/internal/producer"
	"scheduler_service/internal/usecase"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system envs")
	}
	cfg := config.Load()
	logg := logger.NewLogger()

	producer := producer.NewKafkaProducer(cfg.KafkaBrokers,cfg.OutputTopic)
	uc := usecase.NewSchedulerJobUsecase(producer,logg)
	
	kafkaConsumer := consumer.NewKafkaConsumer(
		cfg.KafkaBrokers,
		cfg.InputTopic,
		cfg.ConsumerGroup,
		uc,
		logg,
	)

	log.Println("Scheduler Service started")

	if err := kafkaConsumer.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}