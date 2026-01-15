package interfaces

import "admin_service/internal/domain/entities"

type ApiKeyRepository interface {
	Create(key *entities.ApiKey) error
	Revoke(keyID string) error
}
