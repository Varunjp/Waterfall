CREATE TABLE api_keys (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL,
    key_hash TEXT NOT NULL,
    scopes TEXT[] NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    revoked_at TIMESTAMP
);