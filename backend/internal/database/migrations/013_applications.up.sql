CREATE TABLE applications (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company      VARCHAR(255) NOT NULL,
    title        VARCHAR(255) NOT NULL,
    status       VARCHAR(20)  NOT NULL DEFAULT 'wishlist',
    location     VARCHAR(255),
    work_mode    VARCHAR(20),
    job_url      TEXT,
    source       VARCHAR(100),
    salary       VARCHAR(100),
    applied_date DATE,
    notes        TEXT,
    sort_order   INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_applications_user_id ON applications(user_id);
CREATE INDEX idx_applications_board ON applications(user_id, status, sort_order);
