ALTER TABLE users ADD COLUMN mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE mfa_totp_secrets (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    encrypted_secret BYTEA NOT NULL,
    nonce            BYTEA NOT NULL,
    verified         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE TABLE mfa_passkeys (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id    BYTEA NOT NULL UNIQUE,
    public_key       BYTEA NOT NULL,
    attestation_type VARCHAR(50),
    transport        TEXT[],
    sign_count       BIGINT NOT NULL DEFAULT 0,
    name             VARCHAR(255) NOT NULL,
    aaguid           BYTEA,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at     TIMESTAMPTZ
);
CREATE INDEX idx_mfa_passkeys_user ON mfa_passkeys(user_id);
CREATE INDEX idx_mfa_passkeys_credential ON mfa_passkeys(credential_id);

CREATE TABLE mfa_backup_codes (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash  VARCHAR(128) NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_mfa_backup_codes_user ON mfa_backup_codes(user_id) WHERE used_at IS NULL;

CREATE TABLE webauthn_challenges (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge    BYTEA NOT NULL,
    session_data JSONB NOT NULL,
    operation    VARCHAR(20) NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_webauthn_challenges_user_op ON webauthn_challenges(user_id, operation);
