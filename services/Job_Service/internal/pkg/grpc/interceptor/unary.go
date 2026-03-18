package interceptor

import (
	"context"

	"google.golang.org/grpc"
)

func UnaryErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	)(interface{},error){

		resp,err := handler(ctx,req)
		if err != nil {
			return nil,MapError(err)
		}

		return resp,nil 
	}
}