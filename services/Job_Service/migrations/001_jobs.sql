CREATE TABLE jobs (
    job_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    payload JSON,
    status VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    schedule_at TIMESTAMP,
    retry INT DEFAULT 1,
    max_retry INT DEFAULT 3,
    manual_retry INT 
);