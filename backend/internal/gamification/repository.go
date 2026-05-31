package gamification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides data access for the gamification system.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new gamification repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ---------------------------------------------------------------------------
// Points
// ---------------------------------------------------------------------------

// GetOrCreatePoints returns the user's points record, creating one if absent.
func (r *Repository) GetOrCreatePoints(ctx context.Context, userID uuid.UUID) (*UserPoints, error) {
	var p UserPoints
	err := r.pool.QueryRow(ctx,
		`INSERT INTO user_points (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
		RETURNING user_id, total_points, level, updated_at`,
		userID,
	).Scan(&p.UserID, &p.TotalPoints, &p.Level, &p.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Already exists; fetch it.
			err = r.pool.QueryRow(ctx,
				`SELECT user_id, total_points, level, updated_at
				FROM user_points WHERE user_id = $1`,
				userID,
			).Scan(&p.UserID, &p.TotalPoints, &p.Level, &p.UpdatedAt)
			if err != nil {
				return nil, fmt.Errorf("fetching user points: %w", err)
			}
			return &p, nil
		}
		return nil, fmt.Errorf("upserting user points: %w", err)
	}
	return &p, nil
}

// AddPoints increments the user's total points and recalculates the level.
func (r *Repository) AddPoints(ctx context.Context, userID uuid.UUID, points int) (*UserPoints, error) {
	var p UserPoints
	err := r.pool.QueryRow(ctx,
		`INSERT INTO user_points (user_id, total_points, level, updated_at)
		VALUES ($1, $2, 1, NOW())
		ON CONFLICT (user_id) DO UPDATE
			SET total_points = user_points.total_points + $2,
			    updated_at = NOW()
		RETURNING user_id, total_points, level, updated_at`,
		userID, points,
	).Scan(&p.UserID, &p.TotalPoints, &p.Level, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("adding points: %w", err)
	}

	// Recalculate level from new total.
	newLevel, _ := LevelForPoints(p.TotalPoints)
	if newLevel != p.Level {
		_, err = r.pool.Exec(ctx,
			`UPDATE user_points SET level = $2, updated_at = NOW() WHERE user_id = $1`,
			userID, newLevel,
		)
		if err != nil {
			return nil, fmt.Errorf("updating level: %w", err)
		}
		p.Level = newLevel
	}

	return &p, nil
}

// ---------------------------------------------------------------------------
// Streaks
// ---------------------------------------------------------------------------

// GetOrCreateStreak returns the user's streak record, creating one if absent.
func (r *Repository) GetOrCreateStreak(ctx context.Context, userID uuid.UUID) (*UserStreak, error) {
	var s UserStreak
	err := r.pool.QueryRow(ctx,
		`INSERT INTO user_streaks (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
		RETURNING user_id, current_streak, longest_streak, last_active_on, freeze_used_at, updated_at`,
		userID,
	).Scan(&s.UserID, &s.CurrentStreak, &s.LongestStreak, &s.LastActiveOn, &s.FreezeUsedAt, &s.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = r.pool.QueryRow(ctx,
				`SELECT user_id, current_streak, longest_streak, last_active_on, freeze_used_at, updated_at
				FROM user_streaks WHERE user_id = $1`,
				userID,
			).Scan(&s.UserID, &s.CurrentStreak, &s.LongestStreak, &s.LastActiveOn, &s.FreezeUsedAt, &s.UpdatedAt)
			if err != nil {
				return nil, fmt.Errorf("fetching user streak: %w", err)
			}
			return &s, nil
		}
		return nil, fmt.Errorf("upserting user streak: %w", err)
	}
	return &s, nil
}

// UpdateStreak persists updated streak values.
func (r *Repository) UpdateStreak(ctx context.Context, streak *UserStreak) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE user_streaks
		SET current_streak = $2, longest_streak = $3, last_active_on = $4, freeze_used_at = $5, updated_at = NOW()
		WHERE user_id = $1`,
		streak.UserID, streak.CurrentStreak, streak.LongestStreak, streak.LastActiveOn, streak.FreezeUsedAt,
	)
	if err != nil {
		return fmt.Errorf("updating streak: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Activity log
// ---------------------------------------------------------------------------

// LogActivity inserts a new activity log entry.
func (r *Repository) LogActivity(ctx context.Context, userID uuid.UUID, action string, entityID *uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO activity_log (id, user_id, action, entity_id) VALUES ($1, $2, $3, $4)`,
		uuid.New(), userID, action, entityID,
	)
	if err != nil {
		return fmt.Errorf("logging activity: %w", err)
	}
	return nil
}

// HasRecentActivity checks whether the user has a matching activity_log entry
// after the given timestamp. When entityID is nil the check ignores entity_id.
func (r *Repository) HasRecentActivity(ctx context.Context, userID uuid.UUID, action string, entityID *uuid.UUID, since time.Time) (bool, error) {
	var exists bool
	var err error

	if entityID != nil {
		err = r.pool.QueryRow(ctx,
			`SELECT EXISTS(
				SELECT 1 FROM activity_log
				WHERE user_id = $1 AND action = $2 AND entity_id = $3 AND created_at > $4
			)`,
			userID, action, *entityID, since,
		).Scan(&exists)
	} else {
		err = r.pool.QueryRow(ctx,
			`SELECT EXISTS(
				SELECT 1 FROM activity_log
				WHERE user_id = $1 AND action = $2 AND entity_id IS NULL AND created_at > $3
			)`,
			userID, action, since,
		).Scan(&exists)
	}

	if err != nil {
		return false, fmt.Errorf("checking recent activity: %w", err)
	}
	return exists, nil
}

// GetHeatmap returns aggregated daily activity counts for the past N days.
func (r *Repository) GetHeatmap(ctx context.Context, userID uuid.UUID, days int) ([]HeatmapEntry, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT created_at::date AS date, COUNT(*) AS count
		FROM activity_log
		WHERE user_id = $1 AND created_at >= NOW() - ($2 || ' days')::interval
		GROUP BY date
		ORDER BY date`,
		userID, fmt.Sprintf("%d", days),
	)
	if err != nil {
		return nil, fmt.Errorf("querying heatmap: %w", err)
	}
	defer rows.Close()

	var entries []HeatmapEntry
	for rows.Next() {
		var e HeatmapEntry
		if err := rows.Scan(&e.Date, &e.Count); err != nil {
			return nil, fmt.Errorf("scanning heatmap entry: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating heatmap: %w", err)
	}

	if entries == nil {
		entries = []HeatmapEntry{}
	}
	return entries, nil
}

// ---------------------------------------------------------------------------
// Badges
// ---------------------------------------------------------------------------

// ListBadges returns all badge definitions.
func (r *Repository) ListBadges(ctx context.Context) ([]Badge, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, description, icon, colour, tier, condition_type, condition_config, is_default, created_at
		FROM badges ORDER BY created_at`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing badges: %w", err)
	}
	defer rows.Close()

	var badges []Badge
	for rows.Next() {
		var b Badge
		if err := rows.Scan(
			&b.ID, &b.Name, &b.Description, &b.Icon, &b.Colour,
			&b.Tier, &b.ConditionType, &b.ConditionConfig, &b.IsDefault, &b.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning badge: %w", err)
		}
		badges = append(badges, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating badges: %w", err)
	}

	if badges == nil {
		badges = []Badge{}
	}
	return badges, nil
}

// ListUserBadges returns badges earned by a specific user.
func (r *Repository) ListUserBadges(ctx context.Context, userID uuid.UUID) ([]UserBadge, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, badge_id, awarded_at
		FROM user_badges WHERE user_id = $1 ORDER BY awarded_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing user badges: %w", err)
	}
	defer rows.Close()

	var badges []UserBadge
	for rows.Next() {
		var b UserBadge
		if err := rows.Scan(&b.ID, &b.UserID, &b.BadgeID, &b.AwardedAt); err != nil {
			return nil, fmt.Errorf("scanning user badge: %w", err)
		}
		badges = append(badges, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user badges: %w", err)
	}

	if badges == nil {
		badges = []UserBadge{}
	}
	return badges, nil
}

// ListBadgesWithEarnedStatus returns all badges with per-user earned/awarded_at populated via LEFT JOIN.
func (r *Repository) ListBadgesWithEarnedStatus(ctx context.Context, userID uuid.UUID) ([]Badge, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT b.id, b.name, b.description, b.icon, b.colour, b.tier,
		        b.condition_type, b.condition_config, b.is_default, b.created_at,
		        ub.awarded_at IS NOT NULL AS earned, ub.awarded_at
		FROM badges b
		LEFT JOIN user_badges ub ON ub.badge_id = b.id AND ub.user_id = $1
		ORDER BY b.created_at`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing badges with earned status: %w", err)
	}
	defer rows.Close()

	var badges []Badge
	for rows.Next() {
		var b Badge
		if err := rows.Scan(
			&b.ID, &b.Name, &b.Description, &b.Icon, &b.Colour,
			&b.Tier, &b.ConditionType, &b.ConditionConfig, &b.IsDefault, &b.CreatedAt,
			&b.Earned, &b.AwardedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning badge with earned status: %w", err)
		}
		badges = append(badges, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating badges with earned status: %w", err)
	}

	if badges == nil {
		badges = []Badge{}
	}
	return badges, nil
}

// AwardBadge grants a badge to a user. No-ops on conflict (already awarded).
func (r *Repository) AwardBadge(ctx context.Context, userID, badgeID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_badges (id, user_id, badge_id) VALUES ($1, $2, $3)
		ON CONFLICT (user_id, badge_id) DO NOTHING`,
		uuid.New(), userID, badgeID,
	)
	if err != nil {
		return fmt.Errorf("awarding badge: %w", err)
	}
	return nil
}

// HasBadge checks whether a user has already been awarded a specific badge.
func (r *Repository) HasBadge(ctx context.Context, userID, badgeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_badges WHERE user_id = $1 AND badge_id = $2)`,
		userID, badgeID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking badge: %w", err)
	}
	return exists, nil
}

// CountUserBadges returns the number of badges earned by a user.
func (r *Repository) CountUserBadges(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM user_badges WHERE user_id = $1`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting user badges: %w", err)
	}
	return count, nil
}

// GetEntityCount returns the count of a given entity type for a user.
// Supported entities: win, job, certification, skill, evidence.
func (r *Repository) GetEntityCount(ctx context.Context, userID uuid.UUID, entity string) (int, error) {
	var query string
	switch entity {
	case "win":
		query = `SELECT COUNT(*) FROM wins WHERE user_id = $1`
	case "job":
		query = `SELECT COUNT(*) FROM jobs WHERE user_id = $1`
	case "certification":
		query = `SELECT COUNT(*) FROM certifications WHERE user_id = $1`
	case "skill":
		query = `SELECT COUNT(*) FROM skills WHERE user_id = $1`
	case "evidence":
		query = `SELECT COUNT(*) FROM skill_evidence se
			JOIN skills s ON s.id = se.skill_id WHERE s.user_id = $1`
	case "cert_passed":
		query = `SELECT COUNT(*) FROM certifications WHERE user_id = $1 AND status = 'passed'`
	default:
		return 0, fmt.Errorf("unknown entity type: %s", entity)
	}

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting %s: %w", entity, err)
	}
	return count, nil
}

// ListAllUserIDs returns every user ID in the system.
func (r *Repository) ListAllUserIDs(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `SELECT id FROM users ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("listing user IDs: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning user ID: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user IDs: %w", err)
	}
	return ids, nil
}

// ---------------------------------------------------------------------------
// Badge admin CRUD
// ---------------------------------------------------------------------------

// GetBadgeByName returns a single badge by its name.
func (r *Repository) GetBadgeByName(ctx context.Context, name string) (Badge, error) {
	var b Badge
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, icon, colour, tier, condition_type, condition_config, is_default, created_at
		FROM badges WHERE name = $1`,
		name,
	).Scan(
		&b.ID, &b.Name, &b.Description, &b.Icon, &b.Colour,
		&b.Tier, &b.ConditionType, &b.ConditionConfig, &b.IsDefault, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Badge{}, fmt.Errorf("badge not found: %s", name)
		}
		return Badge{}, fmt.Errorf("querying badge by name: %w", err)
	}
	return b, nil
}

// RevokeBadge removes a badge from a user.
func (r *Repository) RevokeBadge(ctx context.Context, userID, badgeID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM user_badges WHERE user_id = $1 AND badge_id = $2`,
		userID, badgeID,
	)
	if err != nil {
		return fmt.Errorf("revoking badge: %w", err)
	}
	return nil
}

// GetBadgeByID returns a single badge by its ID.
func (r *Repository) GetBadgeByID(ctx context.Context, badgeID uuid.UUID) (Badge, error) {
	var b Badge
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, icon, colour, tier, condition_type, condition_config, is_default, created_at
		FROM badges WHERE id = $1`,
		badgeID,
	).Scan(
		&b.ID, &b.Name, &b.Description, &b.Icon, &b.Colour,
		&b.Tier, &b.ConditionType, &b.ConditionConfig, &b.IsDefault, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Badge{}, fmt.Errorf("badge not found")
		}
		return Badge{}, fmt.Errorf("querying badge: %w", err)
	}
	return b, nil
}

// CreateBadge inserts a new badge definition.
func (r *Repository) CreateBadge(ctx context.Context, b *Badge) error {
	b.ID = uuid.New()
	err := r.pool.QueryRow(ctx,
		`INSERT INTO badges (id, name, description, icon, colour, tier, condition_type, condition_config, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at`,
		b.ID, b.Name, b.Description, b.Icon, b.Colour, b.Tier, b.ConditionType, b.ConditionConfig, b.IsDefault,
	).Scan(&b.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating badge: %w", err)
	}
	return nil
}

// UpdateBadge updates an existing badge definition.
func (r *Repository) UpdateBadge(ctx context.Context, b *Badge) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE badges SET name = $2, description = $3, icon = $4, colour = $5, tier = $6,
		    condition_type = $7, condition_config = $8, is_default = $9
		WHERE id = $1`,
		b.ID, b.Name, b.Description, b.Icon, b.Colour, b.Tier, b.ConditionType, b.ConditionConfig, b.IsDefault,
	)
	if err != nil {
		return fmt.Errorf("updating badge: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("badge not found")
	}
	return nil
}

// DeleteBadge removes a badge definition and cascades to user_badges.
func (r *Repository) DeleteBadge(ctx context.Context, badgeID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM badges WHERE id = $1`,
		badgeID,
	)
	if err != nil {
		return fmt.Errorf("deleting badge: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("badge not found")
	}
	return nil
}
