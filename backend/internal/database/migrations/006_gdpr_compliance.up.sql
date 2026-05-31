-- Email verification fields on users
ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMPTZ;

-- Newsletter/marketing consent
ALTER TABLE users ADD COLUMN newsletter_opt_in BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN newsletter_opt_in_at TIMESTAMPTZ;

-- Auto-verify all existing users to prevent lockout after migration.
-- Only new registrations will require email verification going forward.
UPDATE users SET email_verified = TRUE, email_verified_at = NOW();

-- Email verification tokens
CREATE TABLE email_verification_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);
CREATE INDEX idx_email_verification_tokens_token ON email_verification_tokens(token);

-- Email change requests
CREATE TABLE email_change_requests (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    new_email  VARCHAR(255) NOT NULL,
    token      VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_change_requests_user_id ON email_change_requests(user_id);
CREATE INDEX idx_email_change_requests_token ON email_change_requests(token);
