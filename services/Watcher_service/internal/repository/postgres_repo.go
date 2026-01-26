package repository

import (
	"context"
	"database/sql"
	"time"
	"watcher_service/internal/domain"
)

type jobRepo struct {
	db *sql.DB
}

func NewJobRepository(db *sql.DB) JobRepository {
	return &jobRepo{db: db}
}

// Need improvment in query, need to include more logic

func (r *jobRepo) FetchDueJobs(ctx context.Context, now, until time.Time)([]domain.Job,error) {
	query := `
		SELECT job_id, app_id, type, payload, schedule_at,retry,max_retry
		FROM jobs
		WHERE status = 'SCHEDULED'
		  AND schedule_at BETWEEN $1 AND $2
		FOR UPDATE SKIP LOCKED; 
	`
	rows, err := r.db.QueryContext(ctx,query,now,until)
	if err != nil {
		return nil,err 
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		if err := rows.Scan(
			&j.JobID,
			&j.AppID,
			&j.Type,
			&j.Payload,
			&j.ScheduleAt,
			&j.Retry,
			&j.MaxRetries,
		); err != nil {
			return nil,err 
		}
		jobs = append(jobs, j)
	}

	return jobs, nil 
}

func (r *jobRepo) MarkQueued(ctx context.Context, jobID string) error {
	_,err := r.db.ExecContext(ctx,
		`UPDATE jobs SET status='QUEUED' WHERE job_id=$1`,jobID,
	)
	return err 
}