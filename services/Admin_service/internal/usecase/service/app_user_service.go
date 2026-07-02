package service

import (
	"admin_service/internal/domain/entities"
	"admin_service/internal/domain/enums"
	domainErr "admin_service/internal/domain/errors"
	"admin_service/internal/infrastructure/security"
	"admin_service/internal/pkg/utils"
	"admin_service/internal/pkg/validation"
	repo "admin_service/internal/repository/interfaces"
	redisclient "admin_service/internal/repository/redis"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AppUserService struct {
	repo    repo.AppUserRepository
	otpRepo *redisclient.OTPRepo
	mailer  *utils.Mailer
	secret  string
}

func NewAppUserService(r repo.AppUserRepository, rd *redisclient.OTPRepo, m *utils.Mailer, secret string) *AppUserService {
	return &AppUserService{repo: r, otpRepo: rd, mailer: m, secret: secret}
}

func (s *AppUserService) Create(app_id, email, password, role string) error {
	if !enums.ValidUserRoles[role] {
		return domainErr.ErrInvalidRole
	}

	hash, err := security.Hash(password)
	if err != nil {
		return err
	}

	user := &entities.AppUser{
		AppID:        app_id,
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

func (s *AppUserService) AppLogin(email, password string) (string, error) {

	if !validation.IsVaildEmail(email) {
		return "", domainErr.ErrInvalidCredentials
	}
	appUser, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", domainErr.ErrInvalidCredentials
	}
	if err := security.Compare(appUser.PasswordHash, password); err != nil {
		return "", domainErr.ErrInvalidCredentials
	}
	if appUser.Status != "active" {
		return "", domainErr.ErrUserBlocked
	}
	return security.GenerateJWT(s.secret, appUser.ID, appUser.Role, appUser.AppID)
}

func (s *AppUserService) RequestPasswordReset(ctx context.Context, email string) error {
	_, err := s.repo.FindByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	if !s.otpRepo.CanResend(email) {
		return errors.New("please wait before requesting another OTP")
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		return err
	}

	err = s.otpRepo.StoreOTP(email, otp)

	if err != nil {
		return err
	}

	s.otpRepo.SetCooldown(email)

	return s.mailer.SendOtp(email, otp)
}

func (s *AppUserService) VerifyOtp(ctx context.Context, email, otp string) (string, error) {

	storedotp, attempts, err := s.otpRepo.GetOTP(email)
	if err != nil {
		return "", errors.New("otp expired")
	}

	if attempts >= 5 {
		return "", errors.New("too many attempts")
	}

	if storedotp != otp {
		err := s.otpRepo.IncrementAttempt(email)
		if err != nil {
			return "", err
		}
		return "", errors.New("invalid otp")
	}

	return utils.GenerateResetToken(email)
}

func (s *AppUserService) ResetPassword(ctx context.Context, token, password string) error {

	email, err := utils.ParseResetToken(token)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, email, string(hash))
}

func (s *AppUserService) ListPlans(ctx context.Context) ([]*entities.Plan, error) {

	plans, err := s.repo.ListPlans()

	if err != nil {
		return nil, err
	}

	return plans, nil
}

func (s *AppUserService) BlockUser(ctx context.Context, userId, status string) error {

	err := s.repo.BlockUser(ctx, userId, status)

	if err != nil {
		return err
	}

	return nil
}

func (s *AppUserService) UpdateUser(ctx context.Context, userID, role, passwordHash string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	userID = strings.TrimSpace(userID)
	role = strings.TrimSpace(role)
	passwordHash = strings.TrimSpace(passwordHash)

	if userID == "" {
		return domainErr.ErrMissingUserID
	}

	if role == "" && passwordHash == "" {
		return domainErr.ErrNoFieldsToUpdate
	}

	if role != "" && !enums.ValidUserRoles[role] {
		return fmt.Errorf("%w: %q", domainErr.ErrInvalidRole, role)
	}

	hash, err := security.Hash(passwordHash)
	if err != nil {
		return fmt.Errorf("password hash :%w", err)
	}

	if err := s.repo.UpdateUser(ctx, userID, role, hash); err != nil {
		if errors.Is(err, domainErr.ErrUserNotFound) {
			return err
		}
		if errors.Is(err, domainErr.ErrNoFieldsToUpdate) {
			return err
		}
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

func (s *AppUserService) ListPayments(ctx context.Context, app_id, status string, limit, offset int, startDate, endDate *time.Time) ([]entities.Payment, int, error) {

	if !isValidTime(startDate) {
		startDate = nil
	}

	if !isValidTime(endDate) {
		endDate = nil
	}

	payments, total, err := s.repo.ListPayment(ctx, app_id, status, limit, offset, startDate, endDate)

	if err != nil {
		log.Println("Listpayment error: ", err)
	}

	return payments, total, err
}

func (s *AppUserService) GetInvoice(ctx context.Context, app_id, invoice_id string) ([]byte, error) {

	subID, amount, err := s.repo.GetInvoiceSubscriptionID(ctx, app_id, invoice_id)
	if err != nil {
		log.Println("Get subscription_id in user: ", err)
		return nil, err
	}

	data, err := s.repo.GetSubscriptionDetails(ctx, subID, invoice_id)
	if err != nil {
		log.Println("Get subscription details in user: ", err)
		return nil, err
	}

	data.InvoiceNumber = invoice_id
	data.TotalPaid = float64(amount)

	pdf, err := utils.GeneratePDF(*data)
	if err != nil {
		return nil, err
	}

	return pdf, nil
}

func isValidTime(t *time.Time) bool {
	if t == nil {
		return false
	}
	return !t.IsZero() && t.Unix() != 0
}
