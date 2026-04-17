CREATE TABLE apps (
    app_id UUID PRIMARY KEY,
    app_name VARCHAR(255) NOT NULL,
    app_email VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    tier VARCHAR(20) NOT NULL DEFAULT 'free',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);