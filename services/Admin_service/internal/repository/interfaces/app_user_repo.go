package interfaces

import (
	"admin_service/internal/domain/entities"
	"context"
)

type AppUserRepository interface {
	Create(user *entities.AppUser) error
	FindByApp(appID string) ([]*entities.AppUser, error)
	FindByEmail(email string) (*entities.AppUser,error)
	UpdatePassword(ctx context.Context,email,passhash string) error 
	ListPlans() ([]*entities.Plan,error)
	BlockUser(userID,status string) error 
}
