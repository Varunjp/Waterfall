package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUnauthenticated = status.Error(
		codes.Unauthenticated,
		"authentication required",
	)

	ErrUnauthorized = status.Error(
		codes.PermissionDenied,
		"you do not have permission to perform this action",
	)
)
