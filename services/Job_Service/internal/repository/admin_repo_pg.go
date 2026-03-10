package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type adminrepo struct {
	db *pgxpool.Pool
}

func NewAdminRepo(db *pgxpool.Pool) AdminRepository {
	return &adminrepo{db: db}
}

func (a *adminrepo) GetSubscriptionDetails(ctx context.Context,appID string) (int, int, error) {

	pidquery := `SELECT plan_id
	FROM subscriptions WHERE app_id = $1 AND status = 'ACTIVE'`

	var planId string 

	err := a.db.QueryRow(ctx,pidquery,appID).Scan(&planId)
	if err != nil {
		if errors.Is(err,pgx.ErrNoRows) {
			return -1,-1,errors.New("failed to get plan details,please try again later")
		}
		return -1,-1,err 
	}

	pquery := `SELECT monthly_job_limit FROM plans WHERE plan_id = $1`

	var totalLimit int

	err = a.db.QueryRow(ctx,pquery,planId).Scan(&totalLimit)
	if err != nil {
		if errors.Is(err,pgx.ErrNoRows) {
			return -1,-1,errors.New("Could not fetch monthly limit")
		}
		return -1,-1,err 
	}

	uquery := `SELECT COALESCE(jobs_executed,0) 
	FROM usage_monthly 
	WHERE app_id = $1
	AND month = date_trunc('month',CURRENT_DATE)`

	var monthlyUsage int 
	err = a.db.QueryRow(ctx,uquery,appID).Scan(&monthlyUsage)
	if err != nil {
		return -1,-1,err 
	}

	return totalLimit,monthlyUsage,nil 
}
