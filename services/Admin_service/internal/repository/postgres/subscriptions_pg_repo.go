package postgres

import (
	"admin_service/internal/domain/entities"
	"database/sql"
)

type SubscriptionsPGRepo struct {
	db *sql.DB
}

func NewSubscriptionsRepo(db *sql.DB) *SubscriptionsPGRepo {
	return &SubscriptionsPGRepo{db: db}
}

func (s *SubscriptionsPGRepo) ListPlans() ([]*entities.Plan,error) {
	rows,err := s.db.Query(`
		SELECT plan_id,name,monthly_job_limit,price,created_at
		FROM plans
		WHERE name != 'FREE'`,
	)
	if err != nil {
		return nil,err 
	}
	defer rows.Close()

	var plans []*entities.Plan
	for rows.Next() {
		var p entities.Plan
		err := rows.Scan(
			&p.PlanID,
			&p.Name,
			&p.MonthlyJobLimit,
			&p.Price,
			&p.CreatedAt,
		)
		if err != nil {
			return nil,err 
		}
		plans = append(plans, &p)
	}
	return plans,nil 
}