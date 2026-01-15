package interfaces

import "admin_service/internal/domain/entities"

type AuditRepository interface {
	Log(entry *entities.AuditLog) error
}
