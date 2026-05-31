DROP TABLE IF EXISTS webauthn_challenges;
DROP TABLE IF EXISTS mfa_backup_codes;
DROP TABLE IF EXISTS mfa_passkeys;
DROP TABLE IF EXISTS mfa_totp_secrets;
ALTER TABLE users DROP COLUMN IF EXISTS mfa_enabled;
