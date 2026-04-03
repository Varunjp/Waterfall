package gateway

import (
	"api_gateway/internal/client"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RegisterSchedulerRoutes(r *gin.Engine, schedulerClient *client.SchedulerClient) {
	r.GET("/api/v1/runtime/overview", func(c *gin.Context) {
		appID, err := resolveRuntimeAppID(c)
		if err != nil {
			writeSchedulerError(c, err)
			return
		}

		resp, err := schedulerClient.GetTenantRuntime(
			c.Request.Context(),
			requestAuthorization(c),
			appID,
		)
		if err != nil {
			writeSchedulerError(c, err)
			return
		}

		c.JSON(http.StatusOK, resp)
	})
}

func resolveRuntimeAppID(c *gin.Context) (string, error) {
	rawClaims, exists := c.Get("user")
	if !exists {
		return "", status.Error(codes.Unauthenticated, "missing user context")
	}

	claims, ok := rawClaims.(jwt.MapClaims)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "invalid user context")
	}

	role, _ := claims["role"].(string)
	queryAppID := c.Query("app_id")
	claimAppID, _ := claims["app_id"].(string)

	if role == "platform_admin" {
		if queryAppID == "" {
			return "", status.Error(codes.InvalidArgument, "app_id query parameter is required")
		}
		return queryAppID, nil
	}

	if claimAppID == "" {
		return "", status.Error(codes.PermissionDenied, "missing app_id claim")
	}

	if queryAppID != "" && queryAppID != claimAppID {
		return "", status.Error(codes.PermissionDenied, "cannot access another tenant runtime")
	}

	return claimAppID, nil
}

func requestAuthorization(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		return authHeader
	}

	cookie, err := c.Cookie("token")
	if err != nil || cookie == "" {
		return ""
	}

	return "Bearer " + cookie
}

func writeSchedulerError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	httpStatus := http.StatusBadGateway
	switch st.Code() {
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.FailedPrecondition:
		httpStatus = http.StatusPreconditionFailed
	}

	c.JSON(httpStatus, gin.H{"error": st.Message()})
}
