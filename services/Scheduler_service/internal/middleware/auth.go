package middleware

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const platformAdminRole = "platform_admin"

var publicMethods = map[string]bool{
	"/scheduler.Scheduler/Heartbeat":        true,
	"/scheduler.Scheduler/ReportResult":     true,
	"/scheduler.Scheduler/RegisterWorker":   true,
	"/scheduler.Scheduler/WorkerHeartbeat":  true,
	"/scheduler.Scheduler/UnregisterWorker": true,
}

func APIKeyInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		tokenStr := strings.TrimSpace(authHeader[0])
		if !strings.HasPrefix(tokenStr, "Bearer ") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(
			tokenStr,
			claims,
			func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, status.Error(codes.Unauthenticated, "unexpected signing method")
				}
				return []byte(jwtSecret), nil
			},
		)
		if err != nil || !token.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		if claims.Role == "" {
			return nil, status.Error(codes.PermissionDenied, "missing role claim")
		}

		if claims.AppID == "" && claims.Role != platformAdminRole {
			return nil, status.Error(codes.PermissionDenied, "missing app_id claim")
		}

		ctx = context.WithValue(ctx, ctxAppID, claims.AppID)
		ctx = context.WithValue(ctx, ctxRole, claims.Role)

		return handler(ctx, req)
	}
}
