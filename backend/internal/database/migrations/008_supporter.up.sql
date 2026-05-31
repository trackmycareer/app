-- Supporter status fields on users
ALTER TABLE users ADD COLUMN is_one_time_supporter BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN is_subscriber BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN supporter_since TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN polar_customer_id TEXT;

CREATE INDEX idx_users_polar_customer ON users(polar_customer_id) WHERE polar_customer_id IS NOT NULL;

-- Seed Supporter badge (awarded/revoked directly by webhook handler, not by condition checker)
INSERT INTO badges (name, description, icon, colour, tier, condition_type, condition_config, is_default)
VALUES ('Supporter', 'Supporting the development of trackmy.career', '💜', '#a855f7', 'gold', 'action', '{"action":"supporter"}', true);
