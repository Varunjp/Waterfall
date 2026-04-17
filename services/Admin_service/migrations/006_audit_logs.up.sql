CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    actor_type VARCHAR(30),
    actor_id UUID,
    action VARCHAR(100),
    resource VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now()
);