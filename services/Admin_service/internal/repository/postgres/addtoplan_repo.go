package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"
)

type AddToPlanRepo struct {
	db *sql.DB
}

func NewAddToPlanRepo(db *sql.DB) *AddToPlanRepo {
	return &AddToPlanRepo{db: db}
}

func (a *AddToPlanRepo) AddToDefault(ctx context.Context,appID string) error {
	planId := os.Getenv("FREE_PLAN_ID")
	if planId == "" {
		return errors.New("FREE_PLAN_ID not configured")
	}
	now := time.Now().UTC()
	till := now.AddDate(1,0,0)
	_,err := a.db.ExecContext(ctx,`
		INSERT INTO subscriptions(app_id,plan_id,status,start_date,end_date) VALUES($1,$2,'ACTIVE',$3,$4)
		ON CONFLICT (app_id) DO NOTHING
	`,appID,planId,now,till)

	return err 
}

func (a *AddToPlanRepo) ExtendSubscription(ctx context.Context, planID, appID string, durationMonths int) error {

	var endDate time.Time

	err := a.db.QueryRowContext(ctx,
		`SELECT end_date FROM subscriptions WHERE app_id=$1`,
		appID,
	).Scan(&endDate)
	
	now := time.Now().UTC()
	var newEnd time.Time

	if err == sql.ErrNoRows {
		newEnd = now.AddDate(0, durationMonths, 0)

		_, err = a.db.ExecContext(ctx, `
			INSERT INTO subscriptions(app_id,plan_id,status,start_date,end_date)
			VALUES ($1,$2,'ACTIVE',$3,$4)
		`, appID, planID, now, newEnd)

		return err
	}

	if err != nil {
		return err
	}

	if endDate.After(now) {
		newEnd = endDate.AddDate(0, durationMonths, 0)
	} else {
		newEnd = now.AddDate(0, durationMonths, 0)
	}

	_, err = a.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET plan_id=$2,
			status='ACTIVE',
			start_date=$3,
			end_date=$4
		WHERE app_id=$1
	`, appID, planID, now, newEnd)

	return err
}