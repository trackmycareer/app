package notification

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// reminderInitKey marks, in app_settings, that the reminder engine has run its
// first self-suppressing pass. See Service.RunDailyReminders.
const reminderInitKey = "reminder_engine_initialised"

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a notification. It returns true when a new row was written and
// false when the dedup index already held a matching reminder (ON CONFLICT DO
// NOTHING). The caller uses this to decide whether to send an email.
func (r *Repository) Create(ctx context.Context, n *Notification) (bool, error) {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO notifications
			(id, user_id, type, title, body, related_certification_id, reminder_for_date, read_at, email_sent_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (related_certification_id, type, reminder_for_date)
			WHERE related_certification_id IS NOT NULL
		DO NOTHING
		RETURNING created_at`,
		n.ID, n.UserID, n.Type, n.Title, n.Body, n.RelatedCertificationID,
		n.ReminderForDate, n.ReadAt, n.EmailSentAt,
	).Scan(&n.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("inserting notification: %w", err)
	}
	return true, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, params ListParams) ([]Notification, int, error) {
	where := `WHERE user_id = $1`
	if params.UnreadOnly {
		where += ` AND read_at IS NULL`
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications `+where, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting notifications: %w", err)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(
		`SELECT id, user_id, type, title, body, related_certification_id,
			reminder_for_date, read_at, email_sent_at, created_at
		FROM notifications %s ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		where,
	)

	rows, err := r.pool.Query(ctx, query, userID, limit, params.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing notifications: %w", err)
	}
	defer rows.Close()

	var out []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body,
			&n.RelatedCertificationID, &n.ReminderForDate, &n.ReadAt,
			&n.EmailSentAt, &n.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning notification: %w", err)
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterating notifications: %w", err)
	}

	return out, total, nil
}

// MarkRead marks a single notification read. It is idempotent and user-scoped:
// a missing, already-read, or other-user notification is a no-op, not an error.
func (r *Repository) MarkRead(ctx context.Context, userID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET read_at = NOW()
		WHERE id = $1 AND user_id = $2 AND read_at IS NULL`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("marking notification read: %w", err)
	}
	return nil
}

// MarkAllRead marks every unread notification for a user read and returns the
// number affected.
func (r *Repository) MarkAllRead(ctx context.Context, userID uuid.UUID) (int, error) {
	result, err := r.pool.Exec(ctx,
		`UPDATE notifications SET read_at = NOW()
		WHERE user_id = $1 AND read_at IS NULL`,
		userID,
	)
	if err != nil {
		return 0, fmt.Errorf("marking all notifications read: %w", err)
	}
	return int(result.RowsAffected()), nil
}

func (r *Repository) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting unread notifications: %w", err)
	}
	return count, nil
}

// GetPreferences returns the user's stored preferences, or all-defaults when no
// row exists.
func (r *Repository) GetPreferences(ctx context.Context, userID uuid.UUID) (Preferences, error) {
	var p Preferences
	err := r.pool.QueryRow(ctx,
		`SELECT enabled, remind_90, remind_30, remind_7, channel_email, channel_in_app
		FROM reminder_preferences WHERE user_id = $1`,
		userID,
	).Scan(&p.Enabled, &p.Remind90, &p.Remind30, &p.Remind7, &p.ChannelEmail, &p.ChannelInApp)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DefaultPreferences(), nil
		}
		return Preferences{}, fmt.Errorf("querying reminder preferences: %w", err)
	}
	return p, nil
}

func (r *Repository) UpsertPreferences(ctx context.Context, userID uuid.UUID, p Preferences) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO reminder_preferences
			(user_id, enabled, remind_90, remind_30, remind_7, channel_email, channel_in_app, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			enabled = $2, remind_90 = $3, remind_30 = $4, remind_7 = $5,
			channel_email = $6, channel_in_app = $7, updated_at = NOW()`,
		userID, p.Enabled, p.Remind90, p.Remind30, p.Remind7, p.ChannelEmail, p.ChannelInApp,
	)
	if err != nil {
		return fmt.Errorf("upserting reminder preferences: %w", err)
	}
	return nil
}

// ListDueReminders returns every passed certification within the reminder
// window (90 days before expiry to 30 days after) for a verified user who has
// reminders enabled. The caller selects the most-urgent applicable band per row
// via applicableType; this query only scopes and supplies the data.
func (r *Repository) ListDueReminders(ctx context.Context) ([]DueReminder, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT
			c.id, c.user_id, u.email, u.name, c.name, c.provider, c.expiry_date,
			(c.expiry_date - CURRENT_DATE) AS days_remaining,
			c.credential_url,
			COALESCE(rp.channel_email, TRUE),
			COALESCE(rp.channel_in_app, TRUE),
			COALESCE(rp.remind_90, TRUE),
			COALESCE(rp.remind_30, TRUE),
			COALESCE(rp.remind_7, TRUE)
		FROM certifications c
		JOIN users u ON u.id = c.user_id
		LEFT JOIN reminder_preferences rp ON rp.user_id = c.user_id
		WHERE c.status = 'passed'
			AND c.expiry_date IS NOT NULL
			AND u.email_verified = TRUE
			AND COALESCE(rp.enabled, TRUE) = TRUE
			AND (c.expiry_date - CURRENT_DATE) <= 90
			AND (c.expiry_date - CURRENT_DATE) >= -30`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing due reminders: %w", err)
	}
	defer rows.Close()

	var out []DueReminder
	for rows.Next() {
		var d DueReminder
		if err := rows.Scan(
			&d.CertID, &d.UserID, &d.Email, &d.Name, &d.CertName, &d.Provider,
			&d.ExpiryDate, &d.DaysRemaining, &d.CredentialURL,
			&d.ChannelEmail, &d.ChannelInApp,
			&d.Remind90, &d.Remind30, &d.Remind7,
		); err != nil {
			return nil, fmt.Errorf("scanning due reminder: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating due reminders: %w", err)
	}

	return out, nil
}

func (r *Repository) MarkEmailSent(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET email_sent_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("marking notification email sent: %w", err)
	}
	return nil
}

// IsReminderEngineInitialised reports whether the first self-suppressing pass
// has already run.
func (r *Repository) IsReminderEngineInitialised(ctx context.Context) (bool, error) {
	var value string
	err := r.pool.QueryRow(ctx, `SELECT value FROM app_settings WHERE key = $1`, reminderInitKey).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("querying reminder init flag: %w", err)
	}
	return value == "true", nil
}

func (r *Repository) SetReminderEngineInitialised(ctx context.Context) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO app_settings (key, value, updated_at)
		VALUES ($1, 'true', NOW())
		ON CONFLICT (key) DO UPDATE SET value = 'true', updated_at = NOW()`,
		reminderInitKey,
	)
	if err != nil {
		return fmt.Errorf("setting reminder init flag: %w", err)
	}
	return nil
}
