package interfaces

import "admin_service/internal/domain/entities"

type AppUserRepository interface {
	Create(user *entities.AppUser) error
	FindByApp(appID string) ([]*entities.AppUser, error)
}
