CREATE TABLE IF NOT EXISTS payments (
    invoice_id TEXT PRIMARY KEY,
    subscription_id TEXT NOT NULL,
    app_id  UUID NOT NULL,
    app_name TEXT NOT NULL,
    plan_name TEXT,
    plan_amount BIGINT NOT NULL, 
    amount  BIGINT NOT NULL,
    currency VARCHAR(20) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL,
    paid_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_app_id_paid_at ON payments(app_id,paid_at DESC);