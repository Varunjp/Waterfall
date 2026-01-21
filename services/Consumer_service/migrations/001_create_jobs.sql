CREATE TABLE jobs (
    job_id UUID PRIMARY KEY,
    app_id UUID NOT NULL, 
    type TEXT NOT NULL, 
    payload JSONB,
    status TEXT NOT NULL, 
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL 
)