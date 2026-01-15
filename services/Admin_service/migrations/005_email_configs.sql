CREATE TABLE email_configs (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL,
    provider VARCHAR(50) NOT NULL,
    from_email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    UNIQUE(app_id)
);