package interfaces

import "admin_service/internal/domain/entities"

type AppUserUsecase interface {
	Create(appID, email, password, role string) error
	List(appID string) ([]*entities.AppUser, error)
	Login(email,password string)(string,error)
}
