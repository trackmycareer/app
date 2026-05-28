CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    name          VARCHAR(255) NOT NULL,
    avatar_url    TEXT,
    provider      VARCHAR(50) NOT NULL DEFAULT 'email',
    provider_id   VARCHAR(255),
    is_admin      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_provider ON users(provider, provider_id) WHERE provider_id IS NOT NULL;

CREATE TABLE app_settings (
    key        VARCHAR(100) PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO app_settings (key, value) VALUES
    ('registration_enabled', 'true'),
    ('instance_name', 'trackmy.career');

CREATE TABLE tags (
    id      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name    VARCHAR(100) NOT NULL,
    colour  VARCHAR(7) NOT NULL DEFAULT '#6366f1',
    UNIQUE(user_id, name)
);

CREATE INDEX idx_tags_user_id ON tags(user_id);

CREATE TABLE wins (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    occurred_on DATE NOT NULL DEFAULT CURRENT_DATE,
    category    VARCHAR(50) NOT NULL DEFAULT 'general',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wins_user_id ON wins(user_id);
CREATE INDEX idx_wins_occurred_on ON wins(user_id, occurred_on DESC);
CREATE INDEX idx_wins_category ON wins(user_id, category);

CREATE TABLE win_tags (
    win_id UUID NOT NULL REFERENCES wins(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (win_id, tag_id)
);

CREATE TABLE jobs (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company          VARCHAR(255) NOT NULL,
    title            VARCHAR(255) NOT NULL,
    start_date       DATE NOT NULL,
    end_date         DATE,
    employment_type  VARCHAR(50) NOT NULL DEFAULT 'full_time',
    transition_type  VARCHAR(50),
    location         VARCHAR(255),
    remote           BOOLEAN NOT NULL DEFAULT FALSE,
    responsibilities TEXT,
    notes            TEXT,
    sort_order       INT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_user_id ON jobs(user_id);
CREATE INDEX idx_jobs_timeline ON jobs(user_id, start_date DESC);

CREATE TABLE certifications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    provider        VARCHAR(255) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'planning',
    earned_date     DATE,
    expiry_date     DATE,
    cost            DECIMAL(10, 2),
    currency        VARCHAR(3) DEFAULT 'GBP',
    credential_url  TEXT,
    study_notes     TEXT,
    study_progress  INT DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_certifications_user_id ON certifications(user_id);
CREATE INDEX idx_certifications_status ON certifications(user_id, status);
CREATE INDEX idx_certifications_expiry ON certifications(user_id, expiry_date)
    WHERE expiry_date IS NOT NULL AND status = 'passed';

CREATE TABLE skills (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    category        VARCHAR(100),
    proficiency     INT NOT NULL DEFAULT 1,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX idx_skills_user_id ON skills(user_id);
CREATE INDEX idx_skills_category ON skills(user_id, category);

CREATE TABLE skill_evidence (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    skill_id         UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    evidence_type    VARCHAR(20) NOT NULL,
    evidence_id      UUID NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_skill_evidence_skill ON skill_evidence(skill_id);
CREATE INDEX idx_skill_evidence_target ON skill_evidence(evidence_type, evidence_id);
