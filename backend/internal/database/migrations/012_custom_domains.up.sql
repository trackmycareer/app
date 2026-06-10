CREATE TABLE custom_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain VARCHAR(253) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    cloudflare_hostname_id VARCHAR(64),
    ssl_status VARCHAR(20) DEFAULT 'pending',
    accent_colour VARCHAR(7) DEFAULT '#4F46E5',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMPTZ,
    UNIQUE(domain),
    UNIQUE(user_id)
);
CREATE INDEX idx_custom_domains_domain ON custom_domains(domain);
