package interfaces

import (
	"context"
)

type AddToPlanRepo interface {
	AddToDefault(ctx context.Context, appID string) error
	ExtendSubscription(ctx context.Context, planID, appID string, durationMonths int) error 
}