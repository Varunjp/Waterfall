package interfaces

import "admin_service/internal/domain/entities"

type EmailConfigRepository interface {
	Upsert(cfg *entities.EmailConfig) error
}
