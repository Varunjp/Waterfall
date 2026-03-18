package service

import (
	"admin_service/internal/domain/entities"
	"admin_service/internal/pkg/validation"
	repo "admin_service/internal/repository/interfaces"
	"errors"
	"strings"
)

type PlanService struct {
	repo  repo.PlansRepository
}

func NewPlanService(r repo.PlansRepository) *PlanService {
	return &PlanService{repo: r}
}

func (s *PlanService) CreatePlan(name string,jobLimit int,price float64) error {

	if !validation.IsValidLimit(jobLimit) {
		return errors.New("invalid job limit provided")
	}

	if !validation.IsValidName(name) {
		return errors.New("please provide a proper plan name")
	}

	if !validation.IsValidPrice(price) {
		return errors.New("Price cannot be less than Rs.1")
	}

	newPlan := &entities.Plan{
		Name: name,
		MonthlyJobLimit: jobLimit,
		Price: price,
	}

	err := s.repo.CreatePlan(newPlan)

	return err 
}

func (s *PlanService) ListPlan()([]*entities.Plan,error) {
	return s.repo.GetPlans()
}

func (s *PlanService) UpdatePlans(planId,name string,jobLimit int,price float64,stripePriceID string) (*entities.Plan,error) {
	
	if jobLimit != 0 {
		if !validation.IsValidLimit(jobLimit) {
			return nil,errors.New("invalid job limit provided")
		}
	}

	if price != 0 {
		if !validation.IsValidPrice(price) {
			return nil,errors.New("Price cannot be less than Rs.1")
		}
	}

	stripePriceID = strings.TrimSpace(stripePriceID)

	name = strings.TrimSpace(name)

	updatePlan := &entities.Plan{
		PlanID: planId,
		Name: name,
		MonthlyJobLimit: jobLimit,
		Price: price,
		StripeID: stripePriceID,
	}

	return s.repo.UpdatePlan(updatePlan)
}