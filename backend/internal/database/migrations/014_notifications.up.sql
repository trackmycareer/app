-- In-app notification centre records plus an audit of which reminders were sent.
CREATE TABLE notifications (
    id                       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id                  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type                     VARCHAR(50) NOT NULL,
    title                    VARCHAR(255) NOT NULL,
    body                     TEXT NOT NULL,
    related_certification_id UUID REFERENCES certifications(id) ON DELETE CASCADE,
    reminder_for_date        DATE,
    read_at                  TIMESTAMPTZ,
    email_sent_at            TIMESTAMPTZ,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_created ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_user_unread ON notifications(user_id) WHERE read_at IS NULL;

-- Idempotency anchor for the reminder scheduler. A reminder is uniquely
-- identified by (certification, band, expiry date), so a daily re-run is a
-- no-op via ON CONFLICT DO NOTHING. Snapshotting the expiry date means editing
-- a certification's expiry re-arms the whole reminder cycle automatically.
CREATE UNIQUE INDEX uq_notifications_reminder_dedup
    ON notifications(related_certification_id, type, reminder_for_date)
    WHERE related_certification_id IS NOT NULL;

-- Per-user reminder preferences. One row per user, created lazily on first
-- update. A missing row is treated as all-defaults (reminders on) by the
-- scheduler and the read handler.
CREATE TABLE reminder_preferences (
    user_id        UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    remind_90      BOOLEAN NOT NULL DEFAULT TRUE,
    remind_30      BOOLEAN NOT NULL DEFAULT TRUE,
    remind_7       BOOLEAN NOT NULL DEFAULT TRUE,
    channel_email  BOOLEAN NOT NULL DEFAULT TRUE,
    channel_in_app BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Supports the scheduler's user-agnostic scan for due certifications.
CREATE INDEX idx_certifications_due
    ON certifications(expiry_date)
    WHERE status = 'passed' AND expiry_date IS NOT NULL;
