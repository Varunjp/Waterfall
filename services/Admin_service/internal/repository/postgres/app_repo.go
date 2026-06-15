package postgres

import (
	"admin_service/internal/domain/entities"
	domainerr "admin_service/internal/domain/errors"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

type AppRepo struct {
	db *sql.DB
}

func NewAppRepo(db *sql.DB) *AppRepo {
	return &AppRepo{db}
}

func (r *AppRepo) getPlanByID(planID string) (*entities.Plan, error) {
	if planID == "" {
		return nil, domainerr.ErrPlanIDRequired
	}

	var plan entities.Plan

	err := r.db.QueryRow(`
		SELECT plan_id, name, monthly_job_limit, price, stripe_price_id
		FROM plans
		WHERE plan_id = $1
	`, planID).Scan(
		&plan.PlanID,
		&plan.Name,
		&plan.MonthlyJobLimit,
		&plan.Price,
		&plan.StripeID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerr.ErrPlanNotFound
		}

		return nil, err
	}

	return &plan, nil
}

func (r *AppRepo) Create(app *entities.App) (string, error) {
	plan, err := r.getPlanByID(app.PlanID)
	if err != nil {
		return "", err
	}

	app.Tier = strings.ToLower(plan.Name)

	tx, err := r.db.Begin()
	if err != nil {
		return "", err
	}

	var appID string

	err = tx.QueryRow(`
		INSERT INTO apps (app_name, app_email, status, tier, plan_id, free_limit, free_usage)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING app_id
	`,
		app.AppName,
		app.AppEmail,
		app.Status,
		app.Tier,
		app.PlanID,
		plan.MonthlyJobLimit,
		0,
	).Scan(&appID)

	if err != nil {
		if err := tx.Rollback(); err != nil {
			return "",err
		}
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				switch pqErr.Constraint {
				case "emaill_unique":
					return "", domainerr.ErrAppEmailAlreadyExists
				case "name_unique":
					return "", domainerr.ErrAppNameAlreadyExists
				}
			}
		}

		return "", err
	}

	_, err = tx.Exec(`
		INSERT INTO usage_monthly (app_id, month, jobs_executed)
		VALUES ($1, date_trunc('month', CURRENT_DATE), 0)
		ON CONFLICT (app_id, month) DO NOTHING
	`, appID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return "",err 
		}
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return appID, nil
}

func (r *AppRepo) FindAll() ([]*entities.AppDetails, error) {
	rows, err := r.db.Query(`
		SELECT a.app_id,a.app_name,a.app_email,a.status, p.name As PlanName, s.current_period_end FROM apps a JOIN subscriptions s ON a.app_id = s.app_id JOIN plans p ON p.plan_id = s.plan_id ORDER BY a.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*entities.AppDetails
	for rows.Next() {
		var a entities.AppDetails
		err := rows.Scan(
			&a.AppID,
			&a.AppName,
			&a.AppEmail,
			&a.Status,
			&a.PlanName,
			&a.EndDate,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, &a)
	}
	return apps, nil
}

func (r *AppRepo) UpdateStatus(appID, status string) error {
	_, err := r.db.Exec(`
		UPDATE apps
		SET status=$1
		WHERE app_id=$2
	`, status, appID)
	return err
}

func (r *AppRepo) CreateFirst(user *entities.AppUser) error {

	_, err := r.db.Exec(`
		INSERT INTO app_users
		(app_id, email, password_hash, role, status)
		VALUES ($1,$2,$3,$4,$5)
	`, user.AppID, user.Email, user.PasswordHash, user.Role, user.Status)

	if err != nil {
		return err
	}

	return nil
}

func (r *AppRepo) CreateFreePlan(sub *entities.Subscription) error {

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
		current_period_start = EXCLUDED.current_period_start,
		current_period_end = EXCLUDED.current_period_end,
		updated_at = NOW();
	`

	_, err := r.db.Exec(
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
