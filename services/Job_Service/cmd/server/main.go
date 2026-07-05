package main

import (
	"context"
	"job_service/internal/config"
	"job_service/internal/handler"
	postgresclient "job_service/internal/infra/postgres"
	redisclient "job_service/internal/infra/redis"
	"job_service/internal/logger"
	jobmetrics "job_service/internal/metrics"
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

	logg, err := logger.Newlogger("job-service")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := postgresclient.MustConntect(cfg.DBDSN)
	adminDb := postgresclient.MustConntect(cfg.DBADMINDNS)
	jobRepo := repository.NewJobRepo(db)
	logRepo := repository.NewLogRepo(db)
	adminRepo := repository.NewAdminRepo(adminDb)
	rc, err := redisclient.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		panic(err)
	}
	rr := redisRepo.NewRedisRepo(rc.Client, adminRepo)
	jobMetrics := jobmetrics.NewMetrics()
	go func() {
		if err := jobmetrics.RunServer(ctx, ":"+cfg.Metrics.Port, logg); err != nil {
			log.Printf("metrics server error: %v", err)
		}
	}()

	queueProducer := queue.NewKafkaProducer([]string{cfg.KafkaBrokers[0]}, cfg.KafkaTopic)
	testProducer := producer.NewTestKafkaProducer(cfg.KafkaBrokers, cfg.TestTopic, jobMetrics)
	producer := producer.NewKafkaProducer(cfg.KafkaBrokers, cfg.KafkaTopic, jobMetrics)

	uc := usecase.NewJobUsecase(producer, testProducer, logg, rr, jobMetrics)
	dc := usecase.NewDashboardUsecase(jobRepo, logRepo, queueProducer, rr)
	h := handler.NewJobHandler(uc, *dc)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.APIKeyInterceptor(cfg.JWTKey),
			interceptor.UnaryErrorInterceptor(),
			jobMetrics.UnaryServerInterceptor(),
		),
	)

	jobpb.RegisterJobServiceServer(server, h)

	lis, err := net.Listen("tcp", ":"+cfg.PORT)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Job service running on :", cfg.PORT)
	if err := server.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
