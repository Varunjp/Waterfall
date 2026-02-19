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

func (r *jobRepo) FetchDueJobs(ctx context.Context, now, until time.Time) ([]domain.Job, error) {
	query := `
		SELECT job_id, app_id, type, payload, schedule_at,retry,max_retry
		FROM jobs
		WHERE status = 'SCHEDULED'
		  AND schedule_at BETWEEN $1 AND $2
		FOR UPDATE SKIP LOCKED; 
	`
	rows, err := r.db.QueryContext(ctx, query, now, until)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		jobs = append(jobs, j)
	}

	return jobs, nil
}

func (r *jobRepo) MarkQueued(ctx context.Context, jobID string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE jobs SET status='QUEUED' WHERE job_id=$1`,
		jobID,
	)
	return err
}

func (r *jobRepo) Insert(ctx context.Context, job domain.Job) error {
	query := `
		INSERT INTO jobs(job_id, app_id, type, payload, status, created_at, updated_at, schedule_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (job_id) DO NOTHING;
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		job.JobID,
		job.AppID,
		job.Type,
		job.Payload,
		job.Status,
		job.CreatedAt,
		job.UpdatedAt,
		job.ScheduleAt,
	)

	return err
}

func (r *jobRepo) UpdatePayload(ctx context.Context, jobID, payload string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE jobs SET payload=$1, updated_at=NOW() WHERE job_id=$2`,
		payload,
		jobID,
	)
	return err
}

func (r *jobRepo) UpdateStatus(ctx context.Context, jobID string, status domain.JobStatus) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE jobs SET status=$1, updated_at=NOW() WHERE job_id=$2`,
		status,
		jobID,
	)
	return err
}

func (r *jobRepo) RetryJob(ctx context.Context, jobID string, status domain.JobStatus, retry int) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE jobs SET status=$1,retry=$2, schedule_at=NOW() WHERE job_id=$3`,
		status,
		retry,
		jobID,
	)
	return err
}
