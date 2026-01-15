package service

import (
	"admin_service/internal/infrastructure/security"
	repo "admin_service/internal/repository/interfaces"
)

type AdminService struct {
	repo repo.AdminRepository
	secret string 
}

func NewAdminService(r repo.AdminRepository,secret string) *AdminService {
	return &AdminService{r,secret}
}

func (s *AdminService) Login(email,password string)(string,error) {
	admin,err := s.repo.FindByEmail(email)
	if err != nil {
		return "",err 
	}
	if err := security.Compare(admin.PasswordHash,password); err != nil {
		return "",err 
	}
	return security.GenerateJWT(s.secret,admin.ID,"platform_admin","")
}