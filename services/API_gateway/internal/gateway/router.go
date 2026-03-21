package gateway

import (
	"api_gateway/internal/config"
	"api_gateway/internal/middleware"
	"api_gateway/internal/proto/adminpb"
	"api_gateway/internal/proto/jobpb"
	"context"
	"encoding/json"
	"net/http"

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

	err := jobpb.RegisterJobServiceHandlerFromEndpoint(
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
		ctx,gwMux,cfg.AdminServiceURL,opts,
	)

	if err != nil {
		panic(err)
	}

	err = adminpb.RegisterAppUserServiceHandlerFromEndpoint(
		ctx,gwMux,cfg.AdminServiceURL,opts,
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
		middleware.RateLimitMiddleware(rdb,cfg.RateLimit),
	)

	r.GET("/health",func(c *gin.Context){
		c.JSON(200,gin.H{"status":"ok"})
	})

	r.Any("/api/*any",gin.WrapH(gwMux))

	r.Any("/billing/*path",gin.WrapH(billingProxy))

	return &http.Server{
		Addr: ":"+cfg.Port,
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

	json.NewEncoder(w).Encode(map[string]string{
		"error": s.Message(),
	})
}

func headerMatcher(key string) (string,bool) {
	switch key {
	case "Authorization","x-api-key":
		return key,true 
	default:
		return key,false 
	}
}