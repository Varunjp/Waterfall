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

	"golang.org/x/crypto/bcrypt"
)

type AppUserService struct {
	repo repo.AppUserRepository
	otpRepo *redisclient.OTPRepo
	mailer *utils.Mailer
	secret string 
}

func NewAppUserService(r repo.AppUserRepository,rd *redisclient.OTPRepo,m *utils.Mailer,secret string) *AppUserService {
	return &AppUserService{repo: r,otpRepo: rd,mailer: m,secret: secret}
}

func (s *AppUserService) Create(app_id,email, password, role string) error {
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
		AppID: app_id,
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
	
	if !validation.IsVaildEmail(email) {
		return "",domainErr.ErrInvalidCredentials
	}
	appUser,err := s.repo.FindByEmail(email)
	if err != nil {
		return "",domainErr.ErrInvalidCredentials
	}
	if err := security.Compare(appUser.PasswordHash,password); err != nil {
		return "",domainErr.ErrInvalidCredentials
	}
	return security.GenerateJWT(s.secret,appUser.ID,appUser.Role,appUser.AppID)
}

func (s *AppUserService) RequestPasswordReset(ctx context.Context,email string) error {
	_,err := s.repo.FindByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	if !s.otpRepo.CanResend(email) {
		return errors.New("please wait before requesting another OTP")
	}

	otp,err := utils.GenerateOTP()
	if err != nil {
		return err
	}

	err = s.otpRepo.StoreOTP(email,otp)

	if err != nil {
		return err 
	}

	s.otpRepo.SetCooldown(email)

	return s.mailer.SendOtp(email,otp)
}

func (s *AppUserService) VerifyOtp(ctx context.Context,email,otp string)(string,error) {

	storedotp,attempts,err := s.otpRepo.GetOTP(email)
	if err != nil {
		return "",errors.New("otp expired")
	}

	if attempts >= 5 {
		return "",errors.New("too many attempts")
	}

	if storedotp != otp {
		s.otpRepo.IncrementAttempt(email)
		return "",errors.New("invalid otp")
	}

	return utils.GenerateResetToken(email)
}

func (s *AppUserService) ResetPassword(ctx context.Context,token,password string) error {

	email,err := utils.ParseResetToken(token)
	if err != nil {
		return err 
	}

	hash,err := bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)

	if err != nil {
		return err 
	}

	return s.repo.UpdatePassword(ctx,email,string(hash))
}