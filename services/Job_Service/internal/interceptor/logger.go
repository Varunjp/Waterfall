package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func UnaryLogger(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	)(interface{},error) {
		start := time.Now()
		resp,err := handler(ctx,req)

		log.Info("grpc request",
				zap.String("method",info.FullMethod),
				zap.Duration("duration",time.Since(start)),
				zap.Error(err),
		)

		return resp,err 
	}
}