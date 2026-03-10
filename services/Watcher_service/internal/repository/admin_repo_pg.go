package repository

import (
	"context"
	"database/sql"
)

type adminRepo struct {
	db *sql.DB
}

func NewAdminRepo(db *sql.DB) AdminRepository {
	return &adminRepo{db: db}
}

func (a *adminRepo) UpdateUsage(ctx context.Context,appID string) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	dailyQuery := `
		INSERT INTO usage_daily (app_id, date, jobs_executed)
		VALUES ($1, CURRENT_DATE, 1)
		ON CONFLICT (app_id, date)
		DO UPDATE SET jobs_executed = usage_daily.jobs_executed + 1;
	`
	_, err = tx.ExecContext(ctx, dailyQuery, appID)
	if err != nil {
		return err
	}

	monthlyQuery := `
		INSERT INTO usage_monthly (app_id, month, jobs_executed)
		VALUES ($1, date_trunc('month', CURRENT_DATE), 1)
		ON CONFLICT (app_id, month)
		DO UPDATE SET jobs_executed = usage_monthly.jobs_executed + 1;
	`

	_, err = tx.ExecContext(ctx, monthlyQuery, appID)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}
