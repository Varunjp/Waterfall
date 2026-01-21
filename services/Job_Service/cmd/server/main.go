package main

import (
	"job_service/internal/config"
	"job_service/internal/handler"
	"job_service/internal/logger"
	"job_service/internal/middleware"
	"job_service/internal/producer"
	jobpb "job_service/internal/proto/jobpb"
	"job_service/internal/usecase"
	"log"
	"net"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system envs")
	}
	
	cfg := config.Load()

	logg := logger.Newlogger()

	producer := producer.NewKafkaProducer(cfg.KafkaBrokers,cfg.KafkaTopic)
	uc := usecase.NewJobUsecase(producer,logg)
	h := handler.NewJobHandler(uc)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.APIKeyInterceptor(cfg.JWTKey)),
	)
	
	jobpb.RegisterJobServiceServer(server,h)

	lis, err := net.Listen("tcp",":"+cfg.JWTKey)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Job service running on :",cfg.PORT)
	server.Serve(lis)
}