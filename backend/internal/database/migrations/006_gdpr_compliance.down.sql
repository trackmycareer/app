DROP TABLE IF EXISTS email_change_requests;
DROP TABLE IF EXISTS email_verification_tokens;

ALTER TABLE users DROP COLUMN IF EXISTS newsletter_opt_in_at;
ALTER TABLE users DROP COLUMN IF EXISTS newsletter_opt_in;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
