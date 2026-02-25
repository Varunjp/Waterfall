package service

import (
	"admin_service/internal/domain/entities"
	domainErr "admin_service/internal/domain/errors"
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

func (s *AppService) Register(name,email string) (string,error) {
	
	if !validation.IsVaildEmail(email) {
		return "",ErrInvalidEmail
	}

	if !validation.IsValidName(name) {
		return "",ErrInvaildName
	}

	app := &entities.App{
		AppName: name,
		AppEmail: email,
		Status: "active",
		Tier: "free",
	}

	appID,err := s.repo.Create(app);

	if  err != nil {
		if errors.Is(err, domainErr.ErrAppEmailAlreadyExists) {
			return "",domainErr.ErrAppEmailAlreadyExists
		}
		return "",err 
	}

	// Logic: need to implement super user here
	layout := "02/01/2006"
	pass := name+time.Now().Format(layout)
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
		return "",err 
	}

	return appID,nil 
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