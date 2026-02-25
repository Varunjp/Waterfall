package interfaces

import "admin_service/internal/domain/entities"

type AppUsecase interface {
	Register(name, email string) (string,error)
	List() ([]*entities.App,error)
	Block(appID string) error 
	Unblock(appID string)error 
}