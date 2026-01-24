package handler

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func GrpcUnaryInterceptor(ctx context.Context,req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler)(any,error) {
	start := time.Now()
	resp,err := handler(ctx,req)
	_ = start
	return resp,err 
}