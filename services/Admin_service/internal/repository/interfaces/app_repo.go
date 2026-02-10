package interfaces

import "admin_service/internal/domain/entities"

type AppRepository interface {
	Create(app *entities.App) (string,error) 
	CreateFirst(user *entities.AppUser) error
	FindAll() ([]*entities.App,error)
	UpdateStatus(appID, status string)error 
}