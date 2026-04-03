package postgres

import (
	"admin_service/internal/domain/entities"
	domainerr "admin_service/internal/domain/errors"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type AppRepo struct {
	db *sql.DB
}

func NewAppRepo(db *sql.DB) *AppRepo {
	return &AppRepo{db}
}

func (r *AppRepo) Create(app *entities.App) (string,error) {
	var appID string 

	err := r.db.QueryRow(`
		INSERT INTO apps (app_name, app_email, status, tier, plan_id)
		VALUES ($1, $2, $3, $4,$5)
		RETURNING app_id
	`,
		app.AppName,
		app.AppEmail,
		app.Status,
		app.Tier,
		app.PlanID,
	).Scan(&appID)

	if err != nil {
		var pqErr *pq.Error  

		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				switch pqErr.Constraint {
				case "emaill_unique":
					return "", domainerr.ErrAppEmailAlreadyExists
				case "name_unique":
					return "",domainerr.ErrAppNameAlreadyExists
				}
			}
		}

		return "",err 
	}

	return appID,nil 
}

func (r *AppRepo) FindAll()([]*entities.App,error) {
	rows,err := r.db.Query(`
		SELECT app_id,app_name,app_email,status,tier, created_at, updated_at FROM apps ORDER BY created_at DESC
	`)
	if err != nil {
		return nil,err 
	}
	defer rows.Close()

	var apps []*entities.App
	for rows.Next() {
		var a entities.App
		err := rows.Scan(
			&a.AppID,
			&a.AppName,
			&a.AppEmail,
			&a.Status,
			&a.Tier,
			&a.CreatedAt,
			&a.UpdatedAt,
		)
		if err != nil {
			return nil,err 
		}
		apps = append(apps, &a)
	}
	return apps,nil 
}

func(r *AppRepo) UpdateStatus(appID, status string) error {
	_,err := r.db.Exec(`
		UPDATE apps
		SET status=$1
		WHERE app_id=$2
	`,status,appID)
	return err 
}

func (r *AppRepo)CreateFirst(user *entities.AppUser) error {

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