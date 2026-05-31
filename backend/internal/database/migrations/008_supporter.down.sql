DELETE FROM user_badges WHERE badge_id IN (SELECT id FROM badges WHERE name = 'Supporter' AND is_default = true);
DELETE FROM badges WHERE name = 'Supporter' AND is_default = true;

DROP INDEX IF EXISTS idx_users_polar_customer;

ALTER TABLE users DROP COLUMN IF EXISTS polar_customer_id;
ALTER TABLE users DROP COLUMN IF EXISTS supporter_since;
ALTER TABLE users DROP COLUMN IF EXISTS is_subscriber;
ALTER TABLE users DROP COLUMN IF EXISTS is_one_time_supporter;
