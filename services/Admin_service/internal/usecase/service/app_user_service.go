package service

import (
	"admin_service/internal/domain/entities"
	"admin_service/internal/domain/enums"
	domainErr "admin_service/internal/domain/errors"
	"admin_service/internal/infrastructure/security"
	repo "admin_service/internal/repository/interfaces"
)

type AppUserService struct {
	repo repo.AppUserRepository
}

func NewAppUserService(r repo.AppUserRepository) *AppUserService {
	return &AppUserService{repo: r}
}

func (s *AppUserService) Create(email, password, role string) error {
	if role != enums.RoleSuperAdmin &&
		role != enums.RoleAdmin &&
		role != enums.RoleViewer {
		return domainErr.ErrInvalidRole
	}

	hash, err := security.Hash(password)
	if err != nil {
		return err
	}

	user := &entities.AppUser{
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		Status:       "active",
	}
	return s.repo.Create(user)
}

func (s *AppUserService) List(appID string) ([]*entities.AppUser, error) {
	return s.repo.FindByApp(appID)
}
