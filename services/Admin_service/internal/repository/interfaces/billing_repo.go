package interfaces

import (
	"admin_service/internal/domain/entities"
	"context"
	"time"
)

type BillingRepository interface {
	GetPlanByID(ctx context.Context, planID string) (*entities.Plan, error)
	CreateSubscription(ctx context.Context, sub *entities.Subscription) error
	UpdateAppPlan(ctx context.Context, appID string, planID string) error
	UpdateSubscriptionStatus(ctx context.Context, stripeSubID string, status string) error
	UpdateSubscriptionPeriod(ctx context.Context, stripeSubID string, start time.Time, end time.Time) error
	ResetMonthlyUsage(ctx context.Context, stripeSubID string) error
	BlockAppBilling(ctx context.Context, stripeSubID string) error
	GetSubscription(ctx context.Context,appID string)(*entities.Subscription,error)
}