-- Badge definitions (admin-configurable)
CREATE TABLE badges (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name             VARCHAR(255) NOT NULL,
    description      TEXT,
    icon             VARCHAR(50) NOT NULL DEFAULT '🏆',
    colour           VARCHAR(7) NOT NULL DEFAULT '#6366f1',
    tier             VARCHAR(20) NOT NULL DEFAULT 'bronze',
    condition_type   VARCHAR(20) NOT NULL,
    condition_config JSONB NOT NULL,
    is_default       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User badge awards
CREATE TABLE user_badges (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_id   UUID NOT NULL REFERENCES badges(id) ON DELETE CASCADE,
    awarded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, badge_id)
);

CREATE INDEX idx_user_badges_user ON user_badges(user_id);

-- Streak tracking
CREATE TABLE user_streaks (
    user_id         UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current_streak  INT NOT NULL DEFAULT 0,
    longest_streak  INT NOT NULL DEFAULT 0,
    last_active_on  DATE,
    freeze_used_at  DATE,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Activity log for heatmap + streak
CREATE TABLE activity_log (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action     VARCHAR(50) NOT NULL,
    entity_id  UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_log_user_date ON activity_log(user_id, created_at DESC);

-- Points and levels
CREATE TABLE user_points (
    user_id      UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_points INT NOT NULL DEFAULT 0,
    level        INT NOT NULL DEFAULT 1,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Profile fields on users
ALTER TABLE users ADD COLUMN username VARCHAR(50) UNIQUE;
ALTER TABLE users ADD COLUMN bio TEXT;
ALTER TABLE users ADD COLUMN profile_visibility JSONB NOT NULL DEFAULT '{"jobs":true,"certifications":true,"skills":true,"wins":false,"badges":true}';

-- Seed default badges
INSERT INTO badges (name, description, icon, colour, tier, condition_type, condition_config, is_default) VALUES
    ('First Win', 'Record your first career win', '🎯', '#6366f1', 'bronze', 'action', '{"action":"first_win"}', true),
    ('First Role', 'Log your first job role', '💼', '#6366f1', 'bronze', 'action', '{"action":"first_job"}', true),
    ('First Certification', 'Start tracking a certification', '📜', '#6366f1', 'bronze', 'action', '{"action":"first_cert"}', true),
    ('Certified', 'Pass your first certification', '✅', '#34d399', 'bronze', 'action', '{"action":"first_cert_passed"}', true),
    ('Skill Builder', 'Add your first skill', '⚡', '#6366f1', 'bronze', 'action', '{"action":"first_skill"}', true),
    ('Evidence Collector', 'Link evidence to a skill', '🔗', '#6366f1', 'bronze', 'action', '{"action":"first_evidence_link"}', true),
    ('Win Tracker', 'Record 5 wins', '🏅', '#6366f1', 'bronze', 'count', '{"entity":"win","threshold":5}', true),
    ('Win Champion', 'Record 25 wins', '🥈', '#a855f7', 'silver', 'count', '{"entity":"win","threshold":25}', true),
    ('Win Legend', 'Record 100 wins', '🥇', '#f59e0b', 'gold', 'count', '{"entity":"win","threshold":100}', true),
    ('Cert Collector', 'Track 3 certifications', '📚', '#6366f1', 'bronze', 'count', '{"entity":"certification","threshold":3}', true),
    ('Cert Master', 'Track 10 certifications', '🎓', '#a855f7', 'silver', 'count', '{"entity":"certification","threshold":10}', true),
    ('Skill Mapper', 'Map 5 skills', '🗺️', '#6366f1', 'bronze', 'count', '{"entity":"skill","threshold":5}', true),
    ('Skill Architect', 'Map 20 skills', '🏗️', '#a855f7', 'silver', 'count', '{"entity":"skill","threshold":20}', true),
    ('Week Warrior', 'Maintain a 7-day streak', '🔥', '#f59e0b', 'bronze', 'streak', '{"days":7}', true),
    ('Streak Master', 'Maintain a 30-day streak', '💪', '#f59e0b', 'gold', 'streak', '{"days":30}', true);
