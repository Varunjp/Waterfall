package main

import (
	"job_service/internal/config"
	"job_service/internal/handler"
	postgresclient "job_service/internal/infra/postgres"
	redisclient "job_service/internal/infra/redis"
	"job_service/internal/logger"
	"job_service/internal/middleware"
	"job_service/internal/pkg/grpc/interceptor"
	"job_service/internal/producer"
	jobpb "job_service/internal/proto/jobpb"
	"job_service/internal/queue"
	"job_service/internal/repository"
	redisRepo "job_service/internal/repository/redis"
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
	adminDb := postgresclient.MustConntect(cfg.DBADMINDNS)
	jobRepo := repository.NewJobRepo(db)
	logRepo := repository.NewLogRepo(db)
	adminRepo := repository.NewAdminRepo(adminDb)
	rc,err := redisclient.NewRedisClient(cfg.RedisAddr,cfg.RedisPassword,cfg.RedisDB)
	if err != nil {
		panic(err)
	}
	rr := redisRepo.NewRedisRepo(rc.Client,adminRepo)
	queueProducer := queue.NewKafkaProducer([]string{cfg.KafkaBrokers[0]},cfg.KafkaTopic)

	producer := producer.NewKafkaProducer(cfg.KafkaBrokers,cfg.KafkaTopic)
	uc := usecase.NewJobUsecase(producer,logg,rr)
	dc := usecase.NewDashboardUsecase(jobRepo,logRepo,queueProducer)
	h := handler.NewJobHandler(uc,*dc)

	server := grpc.NewServer(
		// grpc.UnaryInterceptor(middleware.APIKeyInterceptor(cfg.JWTKey)),
		grpc.ChainUnaryInterceptor(
			middleware.APIKeyInterceptor(cfg.JWTKey),
			interceptor.UnaryErrorInterceptor(),
		),
	)
	
	jobpb.RegisterJobServiceServer(server,h)

	lis, err := net.Listen("tcp",":"+cfg.PORT)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Job service running on :",cfg.PORT)
	server.Serve(lis)
}