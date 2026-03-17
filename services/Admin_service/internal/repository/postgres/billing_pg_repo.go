package postgres

import (
	"admin_service/internal/domain/entities"
	"context"
	"database/sql"
	"time"
)

type BillingPGRepo struct {
	db *sql.DB
}

func NewBillingPGRepo(db *sql.DB) *BillingPGRepo {
	return &BillingPGRepo{db: db}
}

func (r *BillingPGRepo) GetPlanByID(ctx context.Context, planID string) (*entities.Plan, error) {

	query := `
		SELECT plan_id,name,monthly_job_limit,price,stripe_price_id
		FROM plans
		WHERE plan_id=$1
	`

	var p entities.Plan

	err := r.db.QueryRowContext(
		ctx,
		query,
		planID,
	).Scan(
		&p.PlanID,
		&p.Name,
		&p.MonthlyJobLimit,
		&p.Price,
		&p.StripeID,
	)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *BillingPGRepo) CreateSubscription(
	ctx context.Context,
	sub *entities.Subscription,
) error {

	query := `
	INSERT INTO subscriptions (
		app_id,
		plan_id,
		stripe_subscription_id,
		status,
		current_period_start,
		current_period_end,
		created_at
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7)
	ON CONFLICT (app_id)
	DO UPDATE SET
		plan_id = EXCLUDED.plan_id,
		updated_at = NOW();
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		sub.AppID,
		sub.PlanID,
		sub.StripeSubscriptionID,
		sub.Status,
		sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd,
		time.Now(),
	)

	return err
}

func (r *BillingPGRepo) UpdateAppPlan(
	ctx context.Context,
	appID string,
	planID string,
) error {

	query := `
	UPDATE apps
	SET plan_id = $1,
	    updated_at = NOW()
	WHERE app_id = $2
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		planID,
		appID,
	)

	return err
}

func (r *BillingPGRepo) UpdateSubscriptionStatus(
	ctx context.Context,
	stripeSubID string,
	status string,
) error {

	query := `
	UPDATE subscriptions
	SET status = $1,
	    updated_at = NOW()
	WHERE stripe_subscription_id = $2
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		status,
		stripeSubID,
	)

	return err
}

func (r *BillingPGRepo) UpdateSubscriptionPeriod(
	ctx context.Context,
	stripeSubID string,
	start time.Time,
	end time.Time,
) error {

	query := `
	UPDATE subscriptions
	SET current_period_start = $1,
	    current_period_end = $2,
	    updated_at = NOW()
	WHERE stripe_subscription_id = $3
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		start,
		end,
		stripeSubID,
	)

	return err
}

func (r *BillingPGRepo) ResetMonthlyUsage(
	ctx context.Context,
	appID string,
) error {

	query := `
	UPDATE usage_monthly
	SET jobs_executed = 0
	WHERE app_id = $1
	AND month = date_trunc('month', CURRENT_DATE)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		appID,
	)

	return err
}

func (r *BillingPGRepo) BlockAppBilling(
	ctx context.Context,
	stripeSubID string,
) error {

	query := `
	UPDATE apps
	SET billing_blocked = TRUE
	WHERE id = (
		SELECT app_id
		FROM subscriptions
		WHERE stripe_subscription_id = $1
	)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		stripeSubID,
	)

	return err
}

func (r *BillingPGRepo)GetSubscription(ctx context.Context,appID string)(*entities.Subscription,error) {

	query := `
		SELECT app_id,plan_id,status,current_period_start,current_period_end,created_at
		FROM subscriptions
		WHERE app_id = $1 AND status = 'ACTIVE' AND current_period_end > NOW();
	`
	var s entities.Subscription
	err := r.db.QueryRowContext(
		ctx,
		query,
		appID,
	).Scan(&s.AppID,&s.PlanID,&s.Status,&s.CurrentPeriodStart,&s.CurrentPeriodEnd,&s.CreatedAt)

	if err != nil {
		return nil,err 
	}

	return &s,nil 
}