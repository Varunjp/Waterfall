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

func (a *adminrepo) GetPlanID(ctx context.Context, appID string) (string, error) {

	pidquery := `SELECT plan_id
	FROM subscriptions 
	WHERE app_id = $1 
	AND LOWER(status) = 'active'
	AND current_period_end > NOW()`

	var planId string

	err := a.db.QueryRow(ctx, pidquery, appID).Scan(&planId)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", errors.New("no active subsription")
	}

	if err != nil {
		return "", err
	}

	return planId, nil
}

func (a *adminrepo) GetPlanDetails(ctx context.Context, planID string) (string, int, error) {
	pquery := `SELECT name,monthly_job_limit FROM plans WHERE plan_id = $1`

	var totalLimit int
	var name string 

	err := a.db.QueryRow(ctx, pquery, planID).Scan(&name,&totalLimit)

	if err != nil {
		return "",-1, err
	}

	return name,totalLimit, nil
}

func (a *adminrepo) GetMonthlyUsage(ctx context.Context, appID string) (int, error) {
	uquery := `SELECT COALESCE((
		SELECT jobs_executed
		FROM usage_monthly
		WHERE app_id = $1
		AND month = date_trunc('month',CURRENT_DATE)
	), 0)`

	var monthlyUsage int
	err := a.db.QueryRow(ctx, uquery, appID).Scan(&monthlyUsage)

	if err != nil {
		return 0, err
	}

	return monthlyUsage, nil
}

func (a *adminrepo) GetFreeQuota(ctx context.Context, appID string) (int, int, error) {
	query := `
		SELECT COALESCE(free_limit, 0), COALESCE(free_usage, 0)
		FROM apps
		WHERE app_id = $1
	`

	var freeLimit, freeUsage int
	err := a.db.QueryRow(ctx, query, appID).Scan(&freeLimit, &freeUsage)

	if err != nil {
		return 0, 0, err
	}

	return freeLimit, freeUsage, nil
}
