package interceptor

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrQuotaExceeded 	= errors.New("quota exceeded")
	ErrNotFound  		= errors.New("not found")
	ErrUnauthorized 	= errors.New("unauthorized")
	ErrInvalidInput 	= errors.New("invalid input")
)

func MapError(err error) error {
	if err == nil {
		return nil 
	}
	
	switch {
	case errors.Is(err,ErrQuotaExceeded):
		return status.Error(codes.ResourceExhausted,err.Error())
	case errors.Is(err, ErrNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, ErrUnauthorized):
		return status.Error(codes.Unauthenticated, err.Error())

	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())

	default:
		return status.Error(codes.Internal, "internal server error")
	}
}