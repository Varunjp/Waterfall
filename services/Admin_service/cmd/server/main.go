package main

import (
	"admin_service/internal/config"
	"admin_service/internal/infrastructure/db"
	pb "admin_service/internal/proto/admin"
	"admin_service/internal/repository/postgres"
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

	usecase := service.NewAdminService(repo,cfg.JWTKey)
	appRepo := postgres.NewAppRepo(dbConn)
	appUsecase := service.NewAppService(appRepo)
	appHandler := handlers.NewAppHandler(appUsecase)
	appUserRepo := postgres.NewAppUserRepo(dbConn)
	appUserUC := service.NewAppUserService(appUserRepo)
	appUserHandler := handlers.NewAppUserHandler(appUserUC)

	apiKeyRepo := postgres.NewApiKeyRepo(dbConn)
	emailRepo := postgres.NewEmailRepo(dbConn)
	//auditRepo := postgres.NewAuditRepo(dbConn)

	apiKeyUC := service.NewApiKeyService(apiKeyRepo)
	emailUC := service.NewEmailService(emailRepo)
	//auditUC := service.NewAuditService(auditRepo)

	handler := handlers.NewAdminHandler(usecase)
	
	lis, _ := net.Listen("tcp",":"+cfg.GrpcPort)
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		interceptors.AuthInterceptor(cfg.JWTKey),
		interceptors.RBACInterceptor(),
	))

	pb.RegisterAdminServiceServer(grpcServer,handler)
	pb.RegisterAppServiceServer(grpcServer, appHandler)
	pb.RegisterAppUserServiceServer(grpcServer, appUserHandler)
	pb.RegisterApiKeyServiceServer(grpcServer, handlers.NewApiKeyHandler(apiKeyUC))
	pb.RegisterEmailServiceServer(grpcServer, handlers.NewEmailHandler(emailUC))

	grpcServer.Serve(lis)
}