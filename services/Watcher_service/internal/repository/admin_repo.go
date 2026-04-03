package repository

import "context"

type AdminRepository interface {
	UpdateUsageIncr(ctx context.Context, appID string) error
	UpdateUsageDecr(ctx context.Context,appID string) error
}