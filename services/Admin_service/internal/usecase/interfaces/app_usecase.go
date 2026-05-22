package interfaces

import "admin_service/internal/domain/entities"

type AppUsecase interface {
	Register(name, email string) (string,string,string,error)
	List() ([]*entities.AppDetails,error)
	Block(appID string) error 
	Unblock(appID string)error 
}