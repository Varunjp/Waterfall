package postgres

import (
	"admin_service/internal/domain/entities"
	"context"
	"database/sql"
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
