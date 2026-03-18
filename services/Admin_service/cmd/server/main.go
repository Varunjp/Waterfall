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
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	_,cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConn := db.NewPostgres(cfg.DBUrl)
	defer dbConn.Close()

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
	defer redisClient.Client.Close()

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
	
	go func(){
		log.Println("admin service listening on",cfg.GrpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Println("grpc stopped:",err)
		}
	}()
	
	billingController := controller.NewBillingController(*billingService,cfg)

	r := chi.NewRouter()

	billingRoutes.RgisterBillingRoutes(
		r,
		billingController,
	)

	httpServer := &http.Server{
		Addr: ":"+cfg.HTTPPort,
		Handler: r,
	}

	go func(){
		log.Println("HTTP billing server running on", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	
	stop := make(chan os.Signal,1)
	signal.Notify(stop,syscall.SIGINT,syscall.SIGTERM)

	<- stop 
	log.Println("Shutting down...")

	cancel()

	grpcServer.GracefulStop()
	ctxShutdown,cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(ctxShutdown); err != nil {
		log.Println("HTTP shutdown error:",err)
	}

	log.Println("Server exited properly")
}