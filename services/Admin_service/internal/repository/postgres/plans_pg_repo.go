package postgres

import (
	"admin_service/internal/domain/entities"
	"database/sql"
	"fmt"
	"strings"
)

type PlanRepo struct {
	db *sql.DB
}

func NewPlanRepo(db *sql.DB) *PlanRepo {
	return &PlanRepo{db: db}
}

func(r *PlanRepo) CreatePlan(plan *entities.Plan)error {

	sname := strings.ToUpper(plan.Name)
	spriceId := "S_"+sname

	_,err := r.db.Exec(`
		INSERT INTO plans(name,monthly_job_limit,price,stripe_price_id)
		VALUES ($1,$2,$3,$4)
	`,plan.Name,plan.MonthlyJobLimit,plan.Price,spriceId)

	return err 
}

func(r *PlanRepo) GetPlans()([]*entities.Plan,error){
	rows,err := r.db.Query(`
		SELECT plan_id,name,monthly_job_limit,price,created_at
		FROM plans`,
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

func(r *PlanRepo)UpdatePlan(plan *entities.Plan)(*entities.Plan,error){
	query := "UPDATE plans SET "
	args := []any{}
	i := 1

	if plan.Name != "" {
		query += fmt.Sprintf("name=$%d,", i)
		args = append(args, plan.Name)
		i++
	}

	if plan.MonthlyJobLimit != 0 {
		query += fmt.Sprintf("monthly_job_limit=$%d,", i)
		args = append(args, plan.MonthlyJobLimit)
		i++
	}

	if plan.Price != 0 {
		query += fmt.Sprintf("price=$%d,", i)
		args = append(args, plan.Price)
		i++
	}

	if plan.StripeID != "" {
		query += fmt.Sprintf("stripe_price_id=$%d",i)
		args = append(args, plan.StripeID)
		i++
	}

	query = strings.TrimSuffix(query, ",")
	query += fmt.Sprintf(" WHERE plan_id=$%d", i)
	args = append(args, plan.PlanID)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil,err
	}

	nquery := `SELECT plan_id,name,monthly_job_limit,price,stripe_price_id FROM plans WHERE plan_id = $1`
	
	var p entities.Plan
	err = r.db.QueryRow(nquery,plan.PlanID).Scan(
		&p.PlanID,&p.Name,&p.MonthlyJobLimit,&p.Price,&p.StripeID,
	)

	if err != nil {
		return nil,err 
	}

	return &p,nil
} 