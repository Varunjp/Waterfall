package interfaces

import (
	"admin_service/internal/domain/entities"
	"context"
	"time"
)

type AdminRepository interface {
	FindByEmail(email string) (*entities.PlatformAdmin, error)
	Create(admin *entities.PlatformAdmin) error
	ListPayment(ctx context.Context, appID, status string, limit, offset int, startDate, endDate *time.Time) ([]entities.Payment, int, error)
	GetPaymentDetails(ctx context.Context, invoiceID string) (*entities.InvoiceData, error)
	ListSubcribers(ctx context.Context, limit, offset int, startDate, endDate *time.Time) ([]entities.Subscriber, int, error)
	GetOverview(ctx context.Context) (*entities.DashboardOverview, error)
}
