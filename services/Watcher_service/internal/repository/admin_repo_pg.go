package repository

import (
	"context"
	"database/sql"
	"log"
)

type adminRepo struct {
	db *sql.DB
}

func NewAdminRepo(db *sql.DB) AdminRepository {
	return &adminRepo{db: db}
}

func (a *adminRepo) UpdateUsageIncr(ctx context.Context, appID string) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("rollback failed :",err)
			}
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

	freeLimit, freeUsage, err := a.getFreeQuota(ctx, tx, appID)
	if err != nil {
		return err
	}

	if freeUsage < freeLimit {
		_, err = tx.ExecContext(ctx, `
			UPDATE apps
			SET free_usage = COALESCE(free_usage, 0) + 1
			WHERE app_id = $1
		`, appID)
		if err != nil {
			return err
		}
	} else {
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
	}

	err = tx.Commit()
	return err
}

func (a *adminRepo) UpdateUsageDecr(ctx context.Context, appID string) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("failed to rollback :",err)
			}
		}
	}()

	dailyQuery := `
		INSERT INTO usage_daily (app_id, date, jobs_executed)
		VALUES ($1, CURRENT_DATE, 0)
		ON CONFLICT (app_id, date)
		DO UPDATE SET jobs_executed = usage_daily.jobs_executed - 1;
	`
	_, err = tx.ExecContext(ctx, dailyQuery, appID)
	if err != nil {
		return err
	}

	monthlyUsage, err := a.getMonthlyUsage(ctx, tx, appID)
	if err != nil {
		return err
	}

	if monthlyUsage > 0 {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO usage_monthly (app_id, month, jobs_executed)
			VALUES ($1, date_trunc('month', CURRENT_DATE), 0)
			ON CONFLICT (app_id, month)
			DO UPDATE SET jobs_executed = GREATEST(usage_monthly.jobs_executed - 1, 0);
		`, appID)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE apps
			SET free_usage = GREATEST(COALESCE(free_usage, 0) - 1, 0)
			WHERE app_id = $1
		`, appID)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (a *adminRepo) getFreeQuota(ctx context.Context, tx *sql.Tx, appID string) (int, int, error) {
	query := `
		SELECT COALESCE(free_limit, 0), COALESCE(free_usage, 0)
		FROM apps
		WHERE app_id = $1
	`

	var freeLimit, freeUsage int
	err := tx.QueryRowContext(ctx, query, appID).Scan(&freeLimit, &freeUsage)
	if err != nil {
		return 0, 0, err
	}

	return freeLimit, freeUsage, nil
}

func (a *adminRepo) getMonthlyUsage(ctx context.Context, tx *sql.Tx, appID string) (int, error) {
	query := `
		SELECT COALESCE((
			SELECT jobs_executed
			FROM usage_monthly
			WHERE app_id = $1
			AND month = date_trunc('month', CURRENT_DATE)
		), 0)
	`

	var monthlyUsage int
	err := tx.QueryRowContext(ctx, query, appID).Scan(&monthlyUsage)
	if err != nil {
		return 0, err
	}

	return monthlyUsage, nil
}
