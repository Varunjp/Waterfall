package service

import (
	"admin_service/internal/domain/entities"
	"admin_service/internal/infrastructure/security"
	"admin_service/internal/pkg/validation"
	repo "admin_service/internal/repository/interfaces"
	"errors"
	"log"
	"time"
)

var (
	ErrInvalidEmail = errors.New("Invalid email provided")
	ErrInvaildName = errors.New("Invalid name provided")
)

type AppService struct {
	repo repo.AppRepository
}

func NewAppService(r repo.AppRepository) *AppService {
	return &AppService{repo: r}
}

func (s *AppService) Register(name,email string) error {
	
	if !validation.IsVaildEmail(email) {
		return ErrInvalidEmail
	}

	if !validation.IsValidName(name) {
		return ErrInvaildName
	}

	app := &entities.App{
		AppName: name,
		AppEmail: email,
		Status: "active",
		Tier: "free",
	}

	appID,err := s.repo.Create(app);

	if  err != nil {
		return err 
	}

	// Logic: need to implement super user here
	pass := name+time.Now().Format("dd/MM/YYYY")
	hash,err := security.Hash(pass)

	if err != nil {
		log.Println(err.Error())
	}

	appUser := &entities.AppUser{
		AppID: appID,
		Email: email,
		PasswordHash: hash,
		Role: "super_admin",
		Status: "active",
	}

	err = s.repo.CreateFirst(appUser)

	if err != nil {
		return err 
	}

	return nil 
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