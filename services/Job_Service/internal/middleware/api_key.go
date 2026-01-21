package middleware

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func APIKeyInterceptor(validKey string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	)(any,error) {
		md,ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil,errors.New("missing metadata")
		}

		keys := md.Get("x-api-key")
		if len(keys) == 0 || keys[0] != validKey {
			return nil , errors.New("unauthorized")
		}

		return handler(ctx,req)
	}
}