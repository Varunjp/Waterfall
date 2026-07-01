ALTER TABLE platform_admins
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

ALTER TABLE apps
    ALTER COLUMN app_id SET DEFAULT gen_random_uuid();

ALTER TABLE app_users
    ALTER COLUMN id SET DEFAULT gen_random_uuid();


ALTER TABLE apps
    ADD COLUMN IF NOT EXISTS plan_id UUID,
    ADD COLUMN IF NOT EXISTS billing_blocked BOOLEAN NOT NULL DEFAULT FALSE;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'emaill_unique'
    ) THEN
        ALTER TABLE apps
            ADD CONSTRAINT emaill_unique UNIQUE (app_email);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'name_unique'
    ) THEN
        ALTER TABLE apps
            ADD CONSTRAINT name_unique UNIQUE (app_name);
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS plans (
    plan_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    monthly_job_limit INT NOT NULL,
    price NUMERIC(12,2) NOT NULL DEFAULT 0,
    stripe_price_id TEXT NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO plans (
    plan_id,
    name,
    monthly_job_limit,
    price,
    stripe_price_id
)
VALUES (
    '501faac9-959d-4311-b8ad-27c8cf951da2',
    'FREE',
    1000000,
    0,
    'S_FREE'
)
ON CONFLICT (plan_id) DO UPDATE
SET
    name = EXCLUDED.name,
    monthly_job_limit = EXCLUDED.monthly_job_limit,
    price = EXCLUDED.price,
    stripe_price_id = EXCLUDED.stripe_price_id;

CREATE TABLE IF NOT EXISTS subscriptions (
    app_id UUID PRIMARY KEY,
    plan_id UUID NOT NULL,
    stripe_subscription_id TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    current_period_start TIMESTAMP NOT NULL DEFAULT NOW(),
    current_period_end TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS subscriptions_stripe_subscription_id_uq
    ON subscriptions (stripe_subscription_id)
    WHERE stripe_subscription_id IS NOT NULL
      AND stripe_subscription_id <> '';

CREATE TABLE IF NOT EXISTS usage_daily (
    app_id UUID NOT NULL,
    date DATE NOT NULL,
    jobs_executed INT NOT NULL DEFAULT 0,
    PRIMARY KEY (app_id, date)
);

CREATE TABLE IF NOT EXISTS usage_monthly (
    app_id UUID NOT NULL,
    month TIMESTAMP NOT NULL,
    jobs_executed INT NOT NULL DEFAULT 0,
    PRIMARY KEY (app_id, month)
);

CREATE TABLE IF NOT EXISTS payments (
    invoice_id TEXT PRIMARY KEY,
    subscription_id TEXT NOT NULL,
    app_id  UUID NOT NULL,
    amount  BIGINT NOT NULL,
    currency VARCHAR(20) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL,
    paid_at TIMESTAMPZ NOT NULL,
    created_at TIMESTAMPZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_app_id_paid_at ON payments(app_id,paid_at DESC);
CREATE INDEX IF NOT EXISTS idx_payments_subscription_id ON payments(subscription_id);