ALTER TABLE jobs
    ALTER COLUMN manual_retry DROP NOT NULL,
    ALTER COLUMN manual_retry DROP DEFAULT;

DROP INDEX IF EXISTS idx_jobs_status_schedule_at;
DROP INDEX IF EXISTS idx_jobs_app_created_at;
DROP INDEX IF EXISTS idx_job_logs_job_id_created_at;
DROP TABLE IF EXISTS job_logs;
