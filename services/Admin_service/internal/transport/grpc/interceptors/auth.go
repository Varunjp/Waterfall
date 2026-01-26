package interceptors

import (
	"context"
	"strings"

	domainErr "admin_service/internal/domain/errors"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func AuthInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func( 
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	)(interface{},error) {

		if publicMethods[info.FullMethod] {
			return handler(ctx,req)
		}

		md,ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil,domainErr.ErrUnauthenticated
		}
		auth := md.Get("authorization")
		if len(auth) == 0 {
			return nil,domainErr.ErrUnauthenticated
		}
		
		tokenStr := strings.TrimPrefix(auth[0],"Bearer ")
		token,err := jwt.Parse(tokenStr,func(t *jwt.Token)(interface{},error){
			return []byte(secret),nil 
		})
		if err != nil || !token.Valid {
			return nil, domainErr.ErrUnauthenticated
		}

		claims,ok := token.Claims.(jwt.MapClaims)
		if !ok  {
			return nil,domainErr.ErrUnauthenticated
		}
		ctx = context.WithValue(ctx,"claims",claims)
		return handler(ctx,req)
	}
}