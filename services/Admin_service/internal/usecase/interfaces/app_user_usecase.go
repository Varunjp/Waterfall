package interfaces

import (
	"admin_service/internal/domain/entities"
	"context"
)

type AppUserUsecase interface {
	Create(appID, email, password, role string) error
	List(appID string) ([]*entities.AppUser, error)
	AppLogin(email,password string)(string,error)
	RequestPasswordReset(ctx context.Context,email string) error
	VerifyOtp(ctx context.Context,email,otp string)(string,error)
	ResetPassword(ctx context.Context,token,password string) error
	ListPlans(ctx context.Context) ([]*entities.Plan,error)
}
