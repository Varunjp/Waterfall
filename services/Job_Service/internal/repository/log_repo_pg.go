package repository

import (
	"context"
	"job_service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type logRepo struct {
	db *pgxpool.Pool
}

func NewLogRepo(db *pgxpool.Pool) JobLogRepository {
	return &logRepo{db: db}
}

func (r *logRepo) GetByJobID(ctx context.Context,jobID, appID string)([]domain.JobLog,error) {

	query := `
	SELECT l.created_at,l.status,l.error
	FROM job_logs l
	JOIN jobs j ON j.job_id = l.job_id
	WHERE l.job_id=$1 AND j.app_id=$2
	ORDER BY l.created_at ASC
	`
	args := []any{jobID,appID}

	rows,err := r.db.Query(ctx,query,args...)

	if err != nil {
		return nil,err 
	}
	defer rows.Close()

	var logs []domain.JobLog
	for rows.Next() {
		var l domain.JobLog
		rows.Scan(&l.Timestamp,&l.Status,&l.ErrorMessage)
		logs = append(logs, l)
	}
	return logs,nil 
}

func (r *logRepo) GetByJobIdAdmin(ctx context.Context,jobID string)([]domain.JobLog,error) {
	query := `
	SELECT l.created_at,l.status,l.error
	FROM job_logs l
	WHERE l.job_id=$1 
	ORDER BY l.created_at ASC
	`
	args := []any{jobID}

	rows,err := r.db.Query(ctx,query,args...)
	if err != nil {
		return nil,err 
	}
	defer rows.Close()

	var logs []domain.JobLog
	for rows.Next() {
		var l domain.JobLog
		rows.Scan(&l.Timestamp,&l.Status,&l.ErrorMessage)
		logs = append(logs, l)
	}
	return logs,nil 
}