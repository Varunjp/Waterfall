package auth

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Unary(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	)(any,error) {

		md,ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil,status.Error(
				codes.Unauthenticated,
				"missing metadata",
			)
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil,status.Error(
				codes.Unauthenticated,
				"missing authorization header",
			)
		}

		tokenStr := strings.TrimSpace(authHeader[0])
		if !strings.HasPrefix(tokenStr,"Bearer "){
			return nil, status.Error(
				codes.Unauthenticated,
				"invalid authorization format",
			)
		}

		tokenStr = strings.TrimPrefix(tokenStr,"Bearer ")

		claims := &Claims{}
		token,err := jwt.ParseWithClaims(
			tokenStr,
			claims,
			func(t *jwt.Token)(any,error) {
				if _,ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil,status.Error(
						codes.Unauthenticated,
						"unexpected signing method",
					)
				}
				return []byte(jwtSecret),nil 
			},
		)

		if err != nil || !token.Valid {
			return nil, status.Error(
				codes.Unauthenticated,
				"invalid or expired token",
			)
		}

		if claims.AppID == "" {
			return nil,status.Error(
				codes.PermissionDenied,
				"missing app_id claim",
			)
		}

		if claims.Role == "" {
			return nil, status.Error(
				codes.PermissionDenied,
				"missing role claim",
			)
		}

		ctx = context.WithValue(ctx, ctxAppID,claims.AppID)
		ctx = context.WithValue(ctx,ctxRole,claims.Role)

		if claims.Subject != "" {
			ctx = context.WithValue(ctx,ctxSub,claims.Subject)
		}

		return handler(ctx,req)
	}
}