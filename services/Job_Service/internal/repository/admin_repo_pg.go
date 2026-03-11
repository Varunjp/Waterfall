package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type adminrepo struct {
	db *pgxpool.Pool
}

func NewAdminRepo(db *pgxpool.Pool) AdminRepository {
	return &adminrepo{db: db}
}

func (a *adminrepo) GetPlanID(ctx context.Context,appID string)(string,error) {
	
	pidquery := `SELECT plan_id
	FROM subscriptions 
	WHERE app_id = $1 
	AND status = 'ACTIVE'
	AND end_date > NOW()`

	var planId string 

	err := a.db.QueryRow(ctx,pidquery,appID).Scan(&planId)

	if err == sql.ErrNoRows {
		return "",errors.New("no active subsription")
	}

	if err != nil  {
		return "",err 
	}

	return planId,nil 
}

func (a *adminrepo) GetPlanDetails(ctx context.Context,planID string)(int,error) {
	pquery := `SELECT monthly_job_limit FROM plans WHERE plan_id = $1`

	var totalLimit int

	err := a.db.QueryRow(ctx,pquery,planID).Scan(&totalLimit)

	if err != nil {
		return -1,err 
	}

	return totalLimit,nil 
}

func (a *adminrepo) GetMonthlyUsage(ctx context.Context,appID string)(int,error) {
	uquery := `SELECT COALESCE(jobs_executed,0) 
	FROM usage_monthly 
	WHERE app_id = $1
	AND month = date_trunc('month',CURRENT_DATE)`

	var monthlyUsage int 
	err := a.db.QueryRow(ctx,uquery,appID).Scan(&monthlyUsage)

	if err != nil {
		return 0,err 
	}

	return monthlyUsage,nil 
}
