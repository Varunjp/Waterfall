package repository

import (
	"context"
	"dashboard_service/internal/domain"
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

	query += " AND status=$2"
	status = strings.ToUpper(status)
	args = append(args, status)
	
	query += " ORDER BY created_at DESC LIMIT $3 OFFSET $4"
	args = append(args,limit,offset)

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

func (r *jobRepo) ListFailed(ctx context.Context, appID string,limit,offset int) ([]domain.Job,error) {
	return r.ListByApp(ctx,appID,"FAILED",limit,offset)
}

func (r *jobRepo) GetByID(ctx context.Context,jobID string)(*domain.Job,error) {
	query := `
	SELECT job_id,app_id, type, payload, status, retry, max_retry, create_at, updated_at
	FROM jobs WHERE job_id=$1
	`
	var j domain.Job
	err := r.db.QueryRow(ctx,query,jobID).Scan(
		&j.JobID,&j.AppID,&j.Type,&j.Payload,&j.Status,&j.Retry,&j.MaxRetry,&j.CreatedAt,&j.UpdatedAt,
	)

	if err != nil {
		return nil,domain.ErrNotFound
	}

	return &j,nil 
}