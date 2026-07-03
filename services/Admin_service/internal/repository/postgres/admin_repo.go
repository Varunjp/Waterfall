package postgres

import (
	"admin_service/internal/domain/entities"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AdminRepo struct {
	db *sql.DB
}

func NewAdminRepo(db *sql.DB) *AdminRepo {
	return &AdminRepo{db}
}

func (r *AdminRepo) FindByEmail(email string) (*entities.PlatformAdmin, error) {
	row := r.db.QueryRow(`
		SELECT id,email,password_hash,status,created_at
		FROM platform_admins WHERE email = $1`, email)

	var padmin entities.PlatformAdmin
	err := row.Scan(&padmin.ID, &padmin.Email, &padmin.PasswordHash, &padmin.Status, &padmin.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &padmin, nil
}

func (r *AdminRepo) Create(admin *entities.PlatformAdmin) error {
	_, err := r.db.Exec(`
		INSERT INTO platform_admins(email,password_hash,status)
		VALUES($1,$2,$3)
	`, admin.Email, admin.PasswordHash, admin.Status)

	return err
}

func (r AdminRepo) ListPayment(ctx context.Context, appID, status string, limit, offset int, startDate, endDate *time.Time) ([]entities.Payment, int, error) {
	baseQuery := `FROM payments WHERE 1=1`
	args := []any{}
	argPos := 1

	// -------- APP FILTER -----------

	if appID != "" {
		baseQuery += fmt.Sprintf(" AND app_id = $%d", argPos)
		args = append(args, appID)
		argPos++
	}

	// -------- STATUS FILTER --------
	if status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, strings.ToLower(status))
		argPos++
	}

	// -------- DATE FILTER --------
	if startDate != nil && endDate != nil {
		baseQuery += fmt.Sprintf(" AND paid_at BETWEEN $%d AND $%d", argPos, argPos+1)
		args = append(args, *startDate, *endDate)
		argPos += 2

	} else if startDate != nil {
		baseQuery += fmt.Sprintf(" AND paid_at >= $%d", argPos)
		args = append(args, *startDate)
		argPos++

	} else if endDate != nil {
		baseQuery += fmt.Sprintf(" AND paid_at <= $%d", argPos)
		args = append(args, *endDate)
		argPos++
	}

	// -------- COUNT QUERY --------
	countQuery := "SELECT COUNT(*) " + baseQuery

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// -------- DATA QUERY --------
	dataQuery := `
	SELECT invoice_id, plan_name, amount, status, paid_at, app_name,customer_email,plan_amount
	` + baseQuery + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	dataArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var payments []entities.Payment
	for rows.Next() {
		var p entities.Payment
		if err := rows.Scan(&p.InvoiceID, &p.PlanName, &p.Amount, &p.Status, &p.PaidAt, &p.AppName, &p.CustomerEmail, &p.PlanAmount); err != nil {
			return nil, 0, err
		}
		payments = append(payments, p)
	}

	return payments, total, nil
}

func (r AdminRepo) ListSubcribers(ctx context.Context, limit, offset int, startDate, endDate *time.Time) ([]entities.Subscriber, int, error) {
	baseQuery := `
	FROM subscriptions s
	INNER JOIN apps a ON s.app_id = a.app_id
	INNER JOIN plans p ON s.plan_id = p.plan_id
	WHERE 1=1 AND p.name != 'FREE'
	`
	args := []any{}
	argPos := 1

	// -------- DATE FILTER --------
	if startDate != nil && endDate != nil {
		baseQuery += fmt.Sprintf(
			" AND s.current_period_start >= $%d AND s.current_period_end <= $%d",
			argPos,
			argPos+1,
		)

		args = append(args, *startDate, *endDate)
		argPos += 2

	} else if startDate != nil {

		baseQuery += fmt.Sprintf(
			" AND s.current_period_start >= $%d",
			argPos,
		)

		args = append(args, *startDate)
		argPos++

	} else if endDate != nil {

		baseQuery += fmt.Sprintf(
			" AND s.current_period_end <= $%d",
			argPos,
		)

		args = append(args, *endDate)
		argPos++
	}

	// -------- COUNT QUERY --------
	countQuery := "SELECT COUNT(*) " + baseQuery

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// -------- DATA QUERY --------
	dataQuery := `
	SELECT
		s.app_id,
		a.app_name,
		s.plan_id,
		p.name,
		s.status,
		s.current_period_start,
		s.current_period_end
	` + baseQuery +
		fmt.Sprintf(
			" ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d",
			argPos,
			argPos+1,
		)

	dataArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	subscribers := make([]entities.Subscriber, 0, limit)

	for rows.Next() {
		var s entities.Subscriber
		if err := rows.Scan(
			&s.AppID,
			&s.AppName,
			&s.PlanID,
			&s.PlanName,
			&s.Status,
			&s.StartDate,
			&s.EndDate,
		); err != nil {
			return nil, 0, err
		}
		subscribers = append(subscribers, s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return subscribers, total, nil
}

func (r AdminRepo) GetPaymentDetails(ctx context.Context, invoiceID string) (*entities.InvoiceData, error) {

	var data entities.InvoiceData

	query := `SELECT invoice_id, app_id, app_name, customer_email, plan_name, plan_amount, amount, paid_at
	FROM payments WHERE invoice_id = $1;`

	err := r.db.QueryRowContext(
		ctx,
		query,
		invoiceID,
	).Scan(
		&data.InvoiceNumber,
		&data.UserID,
		&data.UserName,
		&data.UserEmail,
		&data.PlanName,
		&data.PlanAmount,
		&data.TotalPaid,
		&data.CreatedDate,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("Failed to get invoice: %w", err)
		}
		return nil, err
	}

	return &data, nil
}

func (r AdminRepo) GetOverview(ctx context.Context) (*entities.DashboardOverview,error) {
	query := `
	SELECT

	    /* Users */
	    (SELECT COUNT(*)
	     FROM app_users) AS total_users,

	    /* Apps */
	    (SELECT COUNT(*)
	     FROM apps) AS total_apps,

	    /* Active Subscribers */
	    (SELECT COUNT(*)
	     FROM subscriptions
	     WHERE status='active' AND plan_id != '501faac9-959d-4311-b8ad-27c8cf951da2') AS active_subscribers,

	    /* Revenue Current Month */
	    (
	        SELECT COALESCE(SUM(amount),0)
	        FROM payments
	        WHERE status='paid'
	        AND created_at >= DATE_TRUNC('month', CURRENT_DATE)
	    ) AS revenue_month,

	    /* Revenue Previous Month */
	    (
	        SELECT COALESCE(SUM(amount),0)
	        FROM payments
	        WHERE status='paid'
	        AND created_at >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
	        AND created_at < DATE_TRUNC('month', CURRENT_DATE)
	    ) AS revenue_last_month;
	`

	var overview entities.DashboardOverview

	err := r.db.QueryRowContext(ctx,query).Scan(
		&overview.TotalUsers,
		&overview.TotalApps,
		&overview.ActiveSubscribers,
		&overview.RevenueMonth,
		&overview.RevenueLastMonth,
	)

	if err != nil {
		return nil,err 
	}

	return &overview,nil  
}