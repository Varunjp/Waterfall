CREATE TABLE job_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL,
    attempt INT NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL,
    error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_job_logs_job_id_created_at ON job_logs (job_id, created_at);
CREATE INDEX idx_jobs_app_created_at ON jobs (app_id, created_at DESC);
CREATE INDEX idx_jobs_status_schedule_at ON jobs (status, schedule_at);

ALTER TABLE jobs
    ALTER COLUMN manual_retry SET DEFAULT 0;

UPDATE jobs
SET manual_retry = 0
WHERE manual_retry IS NULL;

ALTER TABLE jobs
    ALTER COLUMN manual_retry SET NOT NULL;
