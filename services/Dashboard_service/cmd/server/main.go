package main

import (
	"dashboard_service/internal/auth"
	"dashboard_service/internal/config"
	postgresclient "dashboard_service/internal/infra/postgres"
	"dashboard_service/internal/queue"
	"dashboard_service/internal/repository"
	pb "dashboard_service/internal/transport/dashboardpb"
	grpcclient "dashboard_service/internal/transport/grpc"
	"dashboard_service/internal/usecase"
	"log"
	"net"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load env")
	}

	cfg := config.Load()
	db := postgresclient.MustConntect(cfg.DBDSN)

	jobRepo := repository.NewJobRepo(db)
	logRepo := repository.NewLogRepo(db)

	queueProducer := queue.NewKafkaProducer([]string{cfg.JobQueueAddr},cfg.Topic)

	uc := usecase.NewDashboardUsecase(jobRepo,logRepo,queueProducer)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(auth.Unary(cfg.JWTSecret)),
	)

	pb.RegisterDashboardServiceServer(
		server,
		grpcclient.NewHandler(uc),
	)

	lis,err := net.Listen("tpc",cfg.ServiceAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("dashboard service listening on",cfg.ServiceAddr)
	server.Serve(lis)
}