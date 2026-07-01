DROP TABLE IF EXISTS usage_monthly;
DROP TABLE IF EXISTS usage_daily;
DROP INDEX IF EXISTS subscriptions_stripe_subscription_id_uq;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS payments;

ALTER TABLE apps
    DROP CONSTRAINT IF EXISTS emaill_unique;

ALTER TABLE apps
    DROP CONSTRAINT IF EXISTS name_unique;

ALTER TABLE apps
    DROP COLUMN IF EXISTS billing_blocked,
    DROP COLUMN IF EXISTS plan_id,
    ALTER COLUMN app_id DROP DEFAULT;

ALTER TABLE platform_admins
    ALTER COLUMN id DROP DEFAULT;

ALTER TABLE app_users
    ALTER COLUMN id DROP DEFAULT;

