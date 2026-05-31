CREATE TABLE linked_accounts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider      VARCHAR(50) NOT NULL,
    provider_id   VARCHAR(255),
    profile_url   TEXT NOT NULL,
    verified      BOOLEAN NOT NULL DEFAULT FALSE,
    verify_token  VARCHAR(255),
    verified_at   TIMESTAMPTZ,
    last_checked  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, provider)
);

CREATE INDEX idx_linked_accounts_user_id ON linked_accounts(user_id);
CREATE INDEX idx_linked_accounts_provider_id ON linked_accounts(provider, provider_id) WHERE provider_id IS NOT NULL;

INSERT INTO linked_accounts (user_id, provider, profile_url, verified, created_at)
SELECT id, 'linkedin', linkedin_url, FALSE, NOW()
FROM users WHERE linkedin_url IS NOT NULL AND linkedin_url != '';

INSERT INTO linked_accounts (user_id, provider, profile_url, verified, created_at)
SELECT id, 'github', github_url, FALSE, NOW()
FROM users WHERE github_url IS NOT NULL AND github_url != '';

INSERT INTO linked_accounts (user_id, provider, profile_url, verified, created_at)
SELECT id, 'website', website_url, FALSE, NOW()
FROM users WHERE website_url IS NOT NULL AND website_url != '';

ALTER TABLE users DROP COLUMN IF EXISTS linkedin_url;
ALTER TABLE users DROP COLUMN IF EXISTS github_url;
ALTER TABLE users DROP COLUMN IF EXISTS website_url;
