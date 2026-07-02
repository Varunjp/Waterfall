package interfaces

import (
	"admin_service/internal/domain/entities"
	"context"
	"time"
)

type AppUserRepository interface {
	Create(user *entities.AppUser) error
	FindByApp(appID string) ([]*entities.AppUser, error)
	FindByEmail(email string) (*entities.AppUser, error)
	UpdatePassword(ctx context.Context, email, passhash string) error
	ListPlans() ([]*entities.Plan, error)
	BlockUser(ctx context.Context, userID, status string) error
	UpdateUser(ctx context.Context, userID, role, passwordHash string) error
	ListPayment(ctx context.Context, appID, status string, limit, offset int, startDate, endDate *time.Time) ([]entities.Payment, int, error)
	GetPaymentDetails(ctx context.Context, appID, invoiceID string) (*entities.InvoiceData, error)
	GetInvoiceSubscriptionID(ctx context.Context, appID, invoiceID string) (string, int64, error)
}
