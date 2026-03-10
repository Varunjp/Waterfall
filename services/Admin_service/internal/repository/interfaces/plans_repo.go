package interfaces

import "admin_service/internal/domain/entities"

type PlansRepository interface {
	CreatePlan(plan *entities.Plan) error 
	GetPlans()([]*entities.Plan,error)
	UpdatePlan(plan *entities.Plan)(*entities.Plan,error) 
}