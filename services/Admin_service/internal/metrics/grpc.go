package metrics

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *AdminMetrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		service, method := splitFullMethod(info.FullMethod)
		endpoint := service + "/" + method
		code := status.Code(err)
		result := resultFromGRPCCode(code)

		m.APIHealth.RequestsTotal.WithLabelValues("grpc", endpoint, code.String()).Inc()
		m.APIHealth.RequestDuration.WithLabelValues("grpc", endpoint, code.String()).Observe(time.Since(start).Seconds())

		if operation := businessOperationForGRPCMethod(method); operation != "" {
			m.BusinessOperations.OperationsTotal.WithLabelValues(operation, "grpc", result).Inc()
			m.BusinessOperations.OperationDuration.WithLabelValues(operation, "grpc", result).Observe(time.Since(start).Seconds())
		}

		m.recordGRPCSecurityEvent(method, endpoint, code)

		return resp, err
	}
}

func resultFromGRPCCode(code codes.Code) string {
	if code == codes.OK {
		return "success"
	}
	return "error"
}

func (m *AdminMetrics) recordGRPCSecurityEvent(method, endpoint string, code codes.Code) {
	switch {
	case method == "Login" || method == "AppLogin":
		m.AuthenticationSecurity.EventsTotal.WithLabelValues("grpc", endpoint, "login", resultFromGRPCCode(code)).Inc()
	case method == "RequestPasswordReset" || method == "VerifyPasswordResetOtp" || method == "ResetPassword":
		m.AuthenticationSecurity.EventsTotal.WithLabelValues("grpc", endpoint, "password_reset", resultFromGRPCCode(code)).Inc()
	case code == codes.Unauthenticated:
		m.AuthenticationSecurity.EventsTotal.WithLabelValues("grpc", endpoint, "authentication", "denied").Inc()
	case code == codes.PermissionDenied:
		m.AuthenticationSecurity.EventsTotal.WithLabelValues("grpc", endpoint, "authorization", "denied").Inc()
	}
}

func splitFullMethod(fullMethod string) (string, string) {
	fullMethod = strings.TrimPrefix(fullMethod, "/")
	parts := strings.Split(fullMethod, "/")
	if len(parts) != 2 {
		return "unknown", fullMethod
	}
	return parts[0], parts[1]
}

func businessOperationForGRPCMethod(method string) string {
	switch method {
	case "RegisterApp", "ListApps", "BlockApp", "UnblockApp":
		return "app_management"
	case "CreateUser", "ListUsers", "UpdateUserStatus":
		return "user_management"
	case "CreatePlan", "ListPlans", "ListPlan", "UpdatePlan":
		return "plan_management"
	default:
		return ""
	}
}
