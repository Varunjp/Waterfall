package main

import (
	"job_service/internal/config"
	"job_service/internal/handler"
	postgresclient "job_service/internal/infra/postgres"
	"job_service/internal/logger"
	"job_service/internal/middleware"
	"job_service/internal/producer"
	jobpb "job_service/internal/proto/jobpb"
	"job_service/internal/queue"
	"job_service/internal/repository"
	"job_service/internal/usecase"
	"log"
	"net"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

// discuss about jwt in job creation

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system envs")
	}
	
	cfg := config.Load()

	logg,err := logger.Newlogger("job-service")
	if err != nil {
		panic(err)
	}
	
	db := postgresclient.MustConntect(cfg.DBDSN)
	jobRepo := repository.NewJobRepo(db)
	logRepo := repository.NewLogRepo(db)
	queueProducer := queue.NewKafkaProducer([]string{cfg.KafkaBrokers[0]},cfg.Topic)

	producer := producer.NewKafkaProducer(cfg.KafkaBrokers,cfg.KafkaTopic)
	uc := usecase.NewJobUsecase(producer,logg)
	dc := usecase.NewDashboardUsecase(jobRepo,logRepo,queueProducer)
	h := handler.NewJobHandler(uc,*dc)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.APIKeyInterceptor(cfg.JWTKey)),
	)
	
	jobpb.RegisterJobServiceServer(server,h)

	lis, err := net.Listen("tcp",":"+cfg.PORT)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Job service running on :",cfg.PORT)
	server.Serve(lis)
}