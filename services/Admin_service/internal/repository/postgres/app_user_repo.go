package postgres

import (
	"admin_service/internal/domain/entities"
	domainErr "admin_service/internal/domain/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AppUserRepo struct {
	db *sql.DB
}

func NewAppUserRepo(db *sql.DB) *AppUserRepo {
	return &AppUserRepo{db}
}

func (r *AppUserRepo) Create(u *entities.AppUser) error {
	_, err := r.db.Exec(`
		INSERT INTO app_users
		(app_id, email, password_hash, role, status)
		VALUES ($1,$2,$3,$4,$5)
	`, u.AppID, u.Email, u.PasswordHash, u.Role, u.Status)

	return err
}

func (r *AppUserRepo) FindByApp(appID string) ([]*entities.AppUser, error) {
	rows, err := r.db.Query(`
		SELECT id, app_id, email, role, status, created_at
		FROM app_users
		WHERE app_id=$1
	`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entities.AppUser
	for rows.Next() {
		var u entities.AppUser
		err := rows.Scan(
			&u.ID,
			&u.AppID,
			&u.Email,
			&u.Role,
			&u.Status,
			&u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

func (r *AppUserRepo) FindByEmail(email string) (*entities.AppUser, error) {
	row := r.db.QueryRow(
		`SELECT id,app_id,email,password_hash,role,status,created_at
		FROM app_users WHERE email = $1`, email,
	)

	var appUser entities.AppUser
	err := row.Scan(&appUser.ID, &appUser.AppID, &appUser.Email, &appUser.PasswordHash, &appUser.Role, &appUser.Status, &appUser.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &appUser, nil
}

func (r *AppUserRepo) UpdatePassword(ctx context.Context, email, passhash string) error {

	query := `UPDATE app_users SET password_hash = $1 WHERE email = $2`
	args := []any{passhash, email}

	_, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *AppUserRepo) ListPlans() ([]*entities.Plan, error) {
	rows, err := r.db.Query(`
		SELECT plan_id,name,monthly_job_limit,price,created_at
		FROM plans
		WHERE name != 'FREE' AND status = 'ACTIVE'`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*entities.Plan
	for rows.Next() {
		var p entities.Plan
		err := rows.Scan(
			&p.PlanID,
			&p.Name,
			&p.MonthlyJobLimit,
			&p.Price,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, &p)
	}
	return plans, nil
}

func (r *AppUserRepo) BlockUser(ctx context.Context, userID, status string) error {

	err := r.db.QueryRowContext(ctx,
		`UPDATE app_users SET status = $1 WHERE id = $2`,
		status,
		userID,
	).Err()

	return err
}

func (r *AppUserRepo) UpdateUser(ctx context.Context, userID, role, passwordHash string) error {

	const selectQuery = `SELECT role,password_hash,status FROM app_users  WHERE id = $1`

	var user entities.AppUser
	err := r.db.QueryRowContext(ctx, selectQuery, userID).Scan(&user.Role, &user.PasswordHash, &user.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainErr.ErrUserNotFound
		}
		return fmt.Errorf("fetch user: %w", err)
	}

	setClauses := make([]string, 0, 2)
	args := make([]any, 0, 3)
	argPos := 1

	if role != "" && role != user.Role {
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", argPos))
		args = append(args, role)
		argPos++
	}

	if passwordHash != "" && passwordHash != user.PasswordHash {
		setClauses = append(setClauses, fmt.Sprintf("password_hash = $%d", argPos))
		args = append(args, passwordHash)
		argPos++
	}

	if len(setClauses) == 0 {
		return domainErr.ErrNoFieldsToUpdate
	}

	args = append(args, userID)

	query := fmt.Sprintf(
		"UPDATE app_users SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "),
		argPos,
	)
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return domainErr.ErrUserNotFound
	}

	return nil
}

func (r *AppUserRepo) ListPayment(ctx context.Context, appID, status string, limit, offset int, startDate, endDate *time.Time) ([]entities.Payment, int, error) {
	baseQuery := `FROM payments WHERE app_id = $1`
	args := []any{appID}
	argPos := 2

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
	SELECT invoice_id, plan_name, amount, status, paid_at
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
		if err := rows.Scan(&p.InvoiceID, &p.PlanName, &p.Amount, &p.Status, &p.PaidAt); err != nil {
			return nil, 0, err
		}
		payments = append(payments, p)
	}

	return payments, total, nil
}

func (r *AppUserRepo) GetPaymentDetails(ctx context.Context, appID, invoiceID string) (*entities.InvoiceData, error) {

	var data entities.InvoiceData

	query := `SELECT invoice_id, app_id, app_name, customer_email, plan_name, plan_amount, amount, paid_at
	FROM payments WHERE invoice_id = $1 AND app_id = $2;`

	err := r.db.QueryRowContext(
		ctx,
		query,
		invoiceID,
		appID,
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

func (r *AppUserRepo) GetInvoiceSubscriptionID(ctx context.Context, appID, invoiceID string) (string, int64, error) {

	query := `SELECT subscription_id,amount FROM payments WHERE invoice_id = $1 AND app_id = $2;`

	var subscriptionID string
	var amount int64
	err := r.db.QueryRowContext(ctx, query, invoiceID, appID).Scan(&subscriptionID, &amount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", 0, fmt.Errorf("Failed to retrive invoice")
		}
		return "", 0, err
	}

	return subscriptionID, amount, nil
}
