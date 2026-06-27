-- Private, encrypted compensation history. One row per compensation event
-- (a raise, a bonus, a new role's package), linked to a job. The monetary
-- amounts live in encrypted_data (an AES-256-GCM JSON blob) so they are never
-- stored or queryable in plaintext. Only the columns needed to sort, group and
-- display the timeline are kept as cleartext.
CREATE TABLE compensation (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id         UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    effective_date DATE NOT NULL,
    currency       VARCHAR(3)  NOT NULL DEFAULT 'GBP',
    pay_basis      VARCHAR(20) NOT NULL DEFAULT 'annual',
    encrypted_data BYTEA NOT NULL,
    nonce          BYTEA NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_compensation_user_id  ON compensation(user_id);
CREATE INDEX idx_compensation_job      ON compensation(user_id, job_id);
CREATE INDEX idx_compensation_timeline ON compensation(user_id, effective_date DESC);
