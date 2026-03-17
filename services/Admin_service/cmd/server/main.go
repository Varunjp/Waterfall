package main

import (
	"admin_service/internal/config"
	"admin_service/internal/infrastructure/db"
	redisclient "admin_service/internal/pkg/redis"
	"admin_service/internal/pkg/utils"
	pb "admin_service/internal/proto/admin"
	"admin_service/internal/repository/postgres"
	redisRepo "admin_service/internal/repository/redis"
	billingRoutes "admin_service/internal/routes"
	"admin_service/internal/transport/grpc/handlers"
	"admin_service/internal/transport/grpc/interceptors"
	controller "admin_service/internal/transport/rest"
	"admin_service/internal/usecase/service"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v78"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("failed to env")
	}
	cfg := config.Load()
	stripe.Key = cfg.Stripe.SecretKey

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
	billingRepo := postgres.NewBillingPGRepo(dbConn)

	appUsecase := service.NewAppService(appRepo,toplanrepo)
	appHandler := handlers.NewAppHandler(appUsecase)
	appUserUC := service.NewAppUserService(appUserRepo,otpRepo,mailer,cfg.JWTKey)
	appUserHandler := handlers.NewAppUserHandler(appUserUC)
	planUC := service.NewPlanService(planRepo)

	type s struct {
		Stripe struct{
			SuccessURL string 
			CancelURL string
		}
	}

	var st s 

	st.Stripe.SuccessURL = cfg.Stripe.SuccessURL
	st.Stripe.CancelURL = cfg.Stripe.CancelURL

	billingService := service.NewBillingService(billingRepo,st,redisClient.Client)

	handler := handlers.NewAdminHandler(usecase,planUC)
	
	go func(){
		lis, err := net.Listen("tcp",":"+cfg.GrpcPort)
		if err != nil {
			log.Fatal("failed to listen :",err)
		}
		grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
			interceptors.AuthInterceptor(cfg.JWTKey),
			interceptors.RBACInterceptor(),
		))

		pb.RegisterAdminServiceServer(grpcServer,handler)
		pb.RegisterAppServiceServer(grpcServer, appHandler)
		pb.RegisterAppUserServiceServer(grpcServer, appUserHandler)

		log.Println("admin service listening on",cfg.GrpcPort)
		grpcServer.Serve(lis)

	}()
	
	billingController := controller.NewBillingController(*billingService,cfg)

	r := chi.NewRouter()

	billingRoutes.RgisterBillingRoutes(
		r,
		billingController,
	)

	log.Println("HTTP billing server running on", cfg.HTTPPort)

	if err := http.ListenAndServe(":"+cfg.HTTPPort, r); err != nil {
		log.Fatal(err)
	}
}