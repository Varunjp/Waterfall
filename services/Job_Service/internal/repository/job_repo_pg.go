package repository

import (
	"context"
	"fmt"
	"job_service/internal/domain"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type jobRepo struct {
	db *pgxpool.Pool
}

func NewJobRepo(db *pgxpool.Pool) JobRepository {
	return &jobRepo{db: db}
}

func (r *jobRepo) ListByApp(ctx context.Context, appID, status string, limit,offset int)([]domain.Job,error) {
	query := `
	SELECT job_id, app_id, type, payload, status, retry, max_retry, created_at,updated_at
	FROM jobs
	WHERE app_id=$1
	`
	
	args := []any{appID}
	argPos := 2

	if status != "" {
		status = strings.ToUpper(status)
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows,err := r.db.Query(ctx,query,args...)
	if err != nil {
		return nil,err 
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		if err := rows.Scan(
			&j.JobID,&j.AppID,&j.Type,&j.Payload,&j.Status,&j.Retry,&j.MaxRetry,&j.CreatedAt,&j.UpdatedAt,
		); err != nil {
			return nil,err 
		}
		jobs = append(jobs, j)
	}

	return jobs,nil 
}

func (r *jobRepo) ListFailed(ctx context.Context,limit,offset int) ([]domain.Job,error) {

	query := `
	SELECT job_id, app_id, type, payload, status, retry, max_retry, created_at,updated_at
	FROM jobs
	WHERE status = 'FAILED' OR status = 'DQL'
	ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	args := []any{}
	args = append(args, limit, offset)

	rows,err := r.db.Query(ctx,query,args...)
	if err != nil {
		return nil,err 
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		if err := rows.Scan(
			&j.JobID,&j.AppID,&j.Type,&j.Payload,&j.Status,&j.Retry,&j.MaxRetry,&j.CreatedAt,&j.UpdatedAt,
		); err != nil {
			return nil,err 
		}
		jobs = append(jobs, j)
	}

	return jobs,nil
}

func (r *jobRepo) GetByID(ctx context.Context,jobID string)(*domain.Job,error) {
	query := `
	SELECT job_id,app_id, type, payload, status, retry, max_retry, created_at, updated_at,manual_retry
	FROM jobs WHERE job_id=$1
	`

	var j domain.Job
	err := r.db.QueryRow(ctx,query,jobID).Scan(
		&j.JobID,&j.AppID,&j.Type,&j.Payload,&j.Status,&j.Retry,&j.MaxRetry,&j.CreatedAt,&j.UpdatedAt,&j.ManualRetry,
	)

	if err != nil {
		return nil,domain.ErrNotFound
	}

	return &j,nil 
}