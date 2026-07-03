package service

import (
	"admin_service/internal/domain/entities"
	domainErr "admin_service/internal/domain/errors"
	"admin_service/internal/infrastructure/security"
	"admin_service/internal/pkg/utils"
	"admin_service/internal/pkg/validation"
	repo "admin_service/internal/repository/interfaces"
	"context"
	"fmt"
	"log"
	"strings"
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

func (s *AdminService) ListPayment(ctx context.Context, appID, status string, limit, offset int, startDate, endDate *time.Time) ([]entities.Payment, int, error) {

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

func (s *AdminService) ListSubcribers(ctx context.Context, limit, offset int, startDate, endDate *time.Time)([]entities.Subscriber,int,error) {

	if !isValidTime(startDate) {
		startDate = nil 
	}

	if !isValidTime(endDate) {
		endDate = nil 
	}

	subscribers,total,err := s.repo.ListSubcribers(ctx,limit,offset,startDate,endDate)

	if err != nil {
		log.Println("List subscribers :",err)
	}

	return subscribers,total,err 
}

func (s *AdminService) GetInvoice(ctx context.Context, invoice_id string) ([]byte, error) {

	invoice_id = strings.TrimSpace(invoice_id)

	if invoice_id == "" {
		return nil, fmt.Errorf("Not valid invoice")
	}

	data, err := s.repo.GetPaymentDetails(ctx, invoice_id)
	if err != nil {
		log.Println("Get subscription details in user: ", err)
		return nil, err
	}

	pdf, err := utils.GeneratePDF(*data)
	if err != nil {
		return nil, err
	}

	return pdf, nil
}

func (s *AdminService) GetOverview(ctx context.Context) (*entities.DashboardOverview,error) {

	overview,err := s.repo.GetOverview(ctx)

	//delete
	fmt.Println("check overview :",overview)
	fmt.Println("Check revenue :",overview.RevenueMonth)

	return overview,err 
}