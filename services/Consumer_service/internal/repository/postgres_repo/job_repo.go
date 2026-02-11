package postgresrepo

import (
	"consumer_service/internal/domain"
	"context"
	"database/sql"
)

type JobRepo struct {
	db *sql.DB
}

func NewJobRepository(db *sql.DB) *JobRepo {
	return &JobRepo{db: db}
}

func (r *JobRepo) Insert(ctx context.Context, job domain.Job) error {
	query := ` 
	INSERT INTO jobs(job_id, app_id, type, payload, status, created_at, updated_at, schedule_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (job_id) DO NOTHING;
	`
	_,err := r.db.ExecContext(ctx, query,
		job.JobID,
		job.AppID,
		job.Type,
		job.Payload,
		job.Status,
		job.CreatedAt,
		job.UpdateAt,
		job.ScheduleAt,
	)

	return err 
}

func (r *JobRepo) UpdatePayload(ctx context.Context, jobID, payload string) error {
	_,err := r.db.ExecContext(ctx,
		`UPDATE jobs SET payload=$1, updated_at=NOW() WHERE job_id=$2`,payload,jobID,
	)
	return err 
}

func (r *JobRepo) UpdateStatus(ctx context.Context, jobID string, status domain.JobStatus) error {
	_,err := r.db.ExecContext(ctx,
		`UPDATE jobs SET status=$1, updated_at=NOW() WHERE job_id=$2`,
		status,jobID,
	)
	return err 
}

func (r *JobRepo) RetryJob(ctx context.Context,jobID string, status domain.JobStatus, retry int) error {
	_,err := r.db.ExecContext(ctx,
		`UPDATE jobs SET status=$1,retry=$2, schedule_at=NOW() WHERE job_id=$3`,
		status,retry,jobID,
	)
	return err 
}