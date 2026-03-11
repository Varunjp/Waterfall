package repository

import "context"

type AdminRepository interface {
	UpdateUsage(ctx context.Context, appID string) error
}