package interfaces

import "admin_service/internal/domain/entities"

type AppRepository interface {
	Create(app *entities.App) error 
	FindAll() ([]*entities.App,error)
	UpdateStatus(appID, status string)error 
}