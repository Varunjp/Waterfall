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
	secret string 
}

func NewAppUserService(r repo.AppUserRepository,secret string) *AppUserService {
	return &AppUserService{repo: r,secret: secret}
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

func (s *AppUserService) AppLogin(email,password string)(string,error) {
	appUser,err := s.repo.FindByEmail(email)
	if err != nil {
		return "",domainErr.ErrInvalidCredentials
	}
	if err := security.Compare(appUser.PasswordHash,password); err != nil {
		return "",domainErr.ErrInvalidCredentials
	}
	return security.GenerateJWT(s.secret,appUser.ID,appUser.Role,appUser.AppID)
}