package service

import (
	"admin_service/internal/domain/entities"
	domainErr "admin_service/internal/domain/errors"
	"admin_service/internal/infrastructure/security"
	"admin_service/internal/pkg/validation"
	repo "admin_service/internal/repository/interfaces"
	"context"
	"log"
	"time"
)

type AdminService struct {
	repo   repo.AdminRepository
	secret string
}

func NewAdminService(r repo.AdminRepository, secret string) *AdminService {
	return &AdminService{r, secret}
}

func (s *AdminService) Login(email, password string) (string, error) {

	if !validation.IsVaildEmail(email) {
		return "", domainErr.ErrInvalidCredentials
	}

	admin, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", err
	}
	if err := security.Compare(admin.PasswordHash, password); err != nil {
		return "", err
	}
	return security.GenerateJWT(s.secret, admin.ID, "platform_admin", "")
}

func (s *AdminService) ListPayment(ctx context.Context,appID, status string, limit, offset int, startDate, endDate *time.Time) ([]entities.Payment, int, error) {
	
	if !isValidTime(startDate) {
		startDate = nil
	}

	if !isValidTime(endDate) {
		endDate = nil
	}

	payments, total, err := s.repo.ListPayment(ctx, appID, status, limit, offset, startDate, endDate)

	if err != nil {
		log.Println("Listpayment error: ", err)
	}

	return payments, total, err
}
