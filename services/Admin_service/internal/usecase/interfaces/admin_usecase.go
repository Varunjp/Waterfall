package interfaces

import (
	"admin_service/internal/domain/entities"
	"context"
	"time"
)

type AdminUsecase interface {
	Login(email, password string) (string, error)
	ListPayment(ctx context.Context,appID, status string, limit, offset int, startDate, endDate *time.Time) ([]entities.Payment, int, error)
}
