package interceptors

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RBACInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	 )(interface{},error) {

		if publicMethods[info.FullMethod] {
			return handler(ctx,req)
		}

		claims,ok := ctx.Value("claims").(jwt.MapClaims)
		if !ok {
			return nil, status.Error(codes.Unauthenticated,"claims missing")
		}

		role,ok := claims["role"].(string)
		if !ok {
			return nil,status.Error(codes.PermissionDenied,"role missing")
		}

		allowedRoles,exists := MethodPermissions[info.FullMethod]
		if !exists {
			return nil, status.Error(codes.PermissionDenied,"method not allowed")
		}

		for _,r := range allowedRoles {
			if r == role {
				return handler(ctx,req)
			}
		}
		return nil,status.Error(codes.PermissionDenied,"access denied")
	 }
}