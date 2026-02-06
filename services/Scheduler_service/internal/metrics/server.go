package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func RunServer(ctx context.Context,addr string,log *zap.Logger) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics",promhttp.Handler())

	server := &http.Server{
		Addr: addr,
		Handler: mux,
	}

	go func() {
		<- ctx.Done()
		log.Info("metrics server shutting down")
		shutdownCtx,cancel := context.WithTimeout(context.Background(), 5 *time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Info("metrics server started", zap.String("addr",addr))
	return server.ListenAndServe()
}