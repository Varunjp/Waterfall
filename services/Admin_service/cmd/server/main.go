package main

import (
	"admin_service/internal/config"
	"admin_service/internal/infrastructure/db"
	redisclient "admin_service/internal/pkg/redis"
	"admin_service/internal/pkg/utils"
	pb "admin_service/internal/proto/admin"
	"admin_service/internal/repository/postgres"
	redisRepo "admin_service/internal/repository/redis"
	"admin_service/internal/transport/grpc/handlers"
	"admin_service/internal/transport/grpc/interceptors"
	"admin_service/internal/usecase/service"
	"log"
	"net"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("failed to env")
	}
	cfg := config.Load()

	dbConn := db.NewPostgres(cfg.DBUrl)
	repo := postgres.NewAdminRepo(dbConn)

	if err := config.BootstrapAdmin(repo); err != nil {
		log.Fatal("admin bootstrap failed:",err)
	}
	redisClient,err := redisclient.NewRedisClient(
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.RedisDB,
	)
	if err != nil {
		log.Fatal("failed to connect redis",err)
	}

	otpRepo := redisRepo.NewOTPRepo(redisClient.Client)
	mailer := utils.NewMailer(
		cfg.SmtpHost,
		cfg.SmtpPort,
		cfg.SmtpUser,
		cfg.Smtppass,
	)
	usecase := service.NewAdminService(repo,cfg.JWTKey)
	appRepo := postgres.NewAppRepo(dbConn)
	appUserRepo := postgres.NewAppUserRepo(dbConn)
	planRepo := postgres.NewPlanRepo(dbConn)
	toplanrepo := postgres.NewAddToPlanRepo(dbConn)
	appUsecase := service.NewAppService(appRepo,toplanrepo)
	appHandler := handlers.NewAppHandler(appUsecase)
	appUserUC := service.NewAppUserService(appUserRepo,otpRepo,mailer,cfg.JWTKey)
	appUserHandler := handlers.NewAppUserHandler(appUserUC)
	planUC := service.NewPlanService(planRepo)

	//auditRepo := postgres.NewAuditRepo(dbConn)

	//auditUC := service.NewAuditService(auditRepo)

	handler := handlers.NewAdminHandler(usecase,planUC)
	
	lis, _ := net.Listen("tcp",":"+cfg.GrpcPort)
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		interceptors.AuthInterceptor(cfg.JWTKey),
		interceptors.RBACInterceptor(),
	))

	pb.RegisterAdminServiceServer(grpcServer,handler)
	pb.RegisterAppServiceServer(grpcServer, appHandler)
	pb.RegisterAppUserServiceServer(grpcServer, appUserHandler)

	log.Println("admin service listening on",cfg.GrpcPort)
	grpcServer.Serve(lis)
}