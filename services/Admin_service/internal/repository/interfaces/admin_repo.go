package interfaces

import "admin_service/internal/domain/entities"

type AdminRepository interface {
	FindByEmail(email string) (*entities.PlatformAdmin,error)
}