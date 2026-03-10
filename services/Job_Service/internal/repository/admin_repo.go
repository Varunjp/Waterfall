package repository

import "context"

type AdminRepository interface {
	GetSubscriptionDetails(ctx context.Context, appID string) (int, int, error)
}