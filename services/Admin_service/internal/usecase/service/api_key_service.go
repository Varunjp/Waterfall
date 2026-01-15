package service

import (
	"admin_service/internal/domain/entities"
	repo "admin_service/internal/repository/interfaces"
	"crypto/rand"
	"encoding/hex"
)

type ApiKeyService struct {
	repo repo.ApiKeyRepository
}

func NewApiKeyService(r repo.ApiKeyRepository) *ApiKeyService {
	return &ApiKeyService{r}
}

func generateKey() (string, string) {
	b := make([]byte, 32)
	rand.Read(b)
	plain := hex.EncodeToString(b)
	return plain, plain 
}

func (s *ApiKeyService) Create(appID string, scopes []string) (string, error) {
	plain, hash := generateKey()
	key := &entities.ApiKey{
		AppID:   appID,
		KeyHash: hash,
		Scopes:  scopes,
	}
	return plain, s.repo.Create(key)
}

func (s *ApiKeyService) Revoke(id string) error {
	return s.repo.Revoke(id)
}
