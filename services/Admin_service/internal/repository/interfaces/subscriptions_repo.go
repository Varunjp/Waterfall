package interfaces

import (
	"admin_service/internal/domain/entities"
)

type SubscriptionsRepo interface {
	ListPlans() ([]*entities.Plan,error)
}