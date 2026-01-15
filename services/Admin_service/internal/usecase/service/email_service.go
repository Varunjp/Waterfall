package service

import (
	"admin_service/internal/domain/entities"
	repo "admin_service/internal/repository/interfaces"
)

type EmailService struct {
	repo repo.EmailConfigRepository
}

func NewEmailService(r repo.EmailConfigRepository) *EmailService {
	return &EmailService{r}
}

func (s *EmailService) Configure(appID, provider, from string) error {
	cfg := &entities.EmailConfig{
		AppID:     appID,
		Provider:  provider,
		FromEmail: from,
	}
	return s.repo.Upsert(cfg)
}
