package repository

import "context"

type AdminRepository interface {
	GetPlanID(ctx context.Context,appID string)(string,error)
	GetPlanDetails(ctx context.Context,planID string)(int,error)
	GetMonthlyUsage(ctx context.Context,appID string)(int,error)
}