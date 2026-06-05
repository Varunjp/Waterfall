package repository

import "context"

type AdminRepository interface {
	GetPlanID(ctx context.Context, appID string) (string, error)
	GetPlanDetails(ctx context.Context, planID string) (string, int, error)
	GetMonthlyUsage(ctx context.Context, appID string) (int, error)
	GetFreeQuota(ctx context.Context, appID string) (int, int, error)
}
