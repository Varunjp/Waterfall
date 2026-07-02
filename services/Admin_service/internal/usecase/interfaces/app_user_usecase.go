package interfaces

import (
	"admin_service/internal/domain/entities"
	"context"
	"time"
)

type AppUserUsecase interface {
	Create(appID, email, password, role string) error
	List(appID string) ([]*entities.AppUser, error)
	AppLogin(email, password string) (string, error)
	RequestPasswordReset(ctx context.Context, email string) error
	VerifyOtp(ctx context.Context, email, otp string) (string, error)
	ResetPassword(ctx context.Context, token, password string) error
	ListPlans(ctx context.Context) ([]*entities.Plan, error)
	BlockUser(ctx context.Context, userId, status string) error
	UpdateUser(ctx context.Context, userID, role, passwordHash string) error
	ListPayments(ctx context.Context, app_id,status string, limit, offset int, startDate, endDate *time.Time)([]entities.Payment,int,error)
	GetInvoice(ctx context.Context, app_id, invoice_id string) ([]byte,error)
}
