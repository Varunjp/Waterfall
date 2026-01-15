package service

import (
	"admin_service/internal/domain/entities"
	repo "admin_service/internal/repository/interfaces"
)

type AppService struct {
	repo repo.AppRepository
}

func NewAppService(r repo.AppRepository) *AppService {
	return &AppService{repo: r}
}

func (s *AppService) Register(name,email string) error {
	app := &entities.App{
		AppName: name,
		AppEmail: email,
		Status: "active",
		Tier: "free",
	}
	return s.repo.Create(app)
}

func (s *AppService) List()([]*entities.App,error) {
	return s.repo.FindAll()
}

func(s *AppService) Block(appID string) error {
	return s.repo.UpdateStatus(appID,"blocked")
}

func(s *AppService) Unblock(appID string) error {
	return s.repo.UpdateStatus(appID,"active")
}