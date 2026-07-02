package gateway

import (
	"api_gateway/internal/config"
	"api_gateway/internal/http/handler"
	"api_gateway/internal/middleware"
	"api_gateway/internal/proto/adminpb"
	"api_gateway/internal/proto/jobpb"
	"api_gateway/internal/proto/schedulerpb"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func NewHTTPServer(ctx context.Context, cfg *config.Config, rdb *redis.Client) *http.Server {

	gwMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(headerMatcher),
		runtime.WithErrorHandler(customErrorHandler),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	adminConn, err := grpc.NewClient(cfg.AdminServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	adminClient := adminpb.NewAppUserServiceClient(adminConn)

	err = jobpb.RegisterJobServiceHandlerFromEndpoint(
		ctx,
		gwMux,
		cfg.JobServiceURL,
		opts,
	)
	if err != nil {
		panic(err)
	}
	err = adminpb.RegisterAdminServiceHandlerFromEndpoint(
		ctx,
		gwMux,
		cfg.AdminServiceURL,
		opts,
	)
	if err != nil {
		panic(err)
	}
	err = adminpb.RegisterAppServiceHandlerFromEndpoint(
		ctx, gwMux, cfg.AdminServiceURL, opts,
	)
	if err != nil {
		panic(err)
	}
	err = adminpb.RegisterAppUserServiceHandlerFromEndpoint(
		ctx, gwMux, cfg.AdminServiceURL, opts,
	)
	if err != nil {
		panic(err)
	}
	err = schedulerpb.RegisterSchedulerHandlerFromEndpoint(
		ctx,
		gwMux,
		cfg.SchedulerServiceURL,
		opts,
	)
	if err != nil {
		panic(err)
	}

	billingProxy := NewReverseProxy(cfg.BillingServiceURL)

	r := gin.New()

	r.Use(
		middleware.RecoveryMiddleware(),
		middleware.LoggingMiddleware(),
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.NoCacheMiddleware(),
		middleware.RateLimitMiddleware(rdb, cfg.RateLimit),
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	invoiceHandler := handler.NewInvoiceHandler(adminClient)

	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/invoice/") {
			invoiceHandler.DownloadInvoice(c)
			c.Abort()
			return
		}
		c.Next()
	})

	r.Any("/api/*any", gin.WrapH(gwMux))

	r.Any("/billing/*path", gin.WrapH(billingProxy))

	RegisterUIRoutes(r)

	return &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}
}

func customErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	m runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	s, _ := status.FromError(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"error": s.Message(),
	}); err != nil {
		log.Println("error :", err)
	}
}

func headerMatcher(key string) (string, bool) {
	switch key {
	case "Authorization", "x-api-key":
		return key, true
	default:
		return key, false
	}
}
