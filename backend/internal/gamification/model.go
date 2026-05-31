package gamification

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/types"
)

// Badge represents a badge definition, optionally enriched with per-user earned status.
type Badge struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Description     *string         `json:"description,omitempty"`
	Icon            string          `json:"icon"`
	Colour          string          `json:"colour"`
	Tier            string          `json:"tier"`
	ConditionType   string          `json:"condition_type"`
	ConditionConfig json.RawMessage `json:"condition_config"`
	IsDefault       bool            `json:"is_default"`
	CreatedAt       time.Time       `json:"created_at"`
	// Populated per-user when listing
	Earned    bool       `json:"earned"`
	AwardedAt *time.Time `json:"awarded_at,omitempty"`
}

// UserBadge records a badge awarded to a specific user.
type UserBadge struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"-"`
	BadgeID   uuid.UUID `json:"badge_id"`
	AwardedAt time.Time `json:"awarded_at"`
}

// UserStreak tracks daily activity streaks for a user.
type UserStreak struct {
	UserID        uuid.UUID   `json:"user_id"`
	CurrentStreak int         `json:"current_streak"`
	LongestStreak int         `json:"longest_streak"`
	LastActiveOn  *types.Date `json:"last_active_on,omitempty"`
	FreezeUsedAt  *types.Date `json:"freeze_used_at,omitempty"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// ActivityLog records a single user action for heatmap and streak calculation.
type ActivityLog struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"-"`
	Action    string     `json:"action"`
	EntityID  *uuid.UUID `json:"entity_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// UserPoints tracks cumulative points and current level for a user.
type UserPoints struct {
	UserID      uuid.UUID `json:"user_id"`
	TotalPoints int       `json:"total_points"`
	Level       int       `json:"level"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// HeatmapEntry represents activity count for a single day.
type HeatmapEntry struct {
	Date  types.Date `json:"date"`
	Count int        `json:"count"`
}

// Progress is the aggregated gamification overview for a user.
type Progress struct {
	TotalPoints   int    `json:"total_points"`
	Level         int    `json:"level"`
	LevelTitle    string `json:"level_title"`
	NextLevel     int    `json:"next_level"`
	NextLevelAt   int    `json:"next_level_at"`
	BadgesEarned  int    `json:"badges_earned"`
	CurrentStreak int    `json:"current_streak"`
	LongestStreak int    `json:"longest_streak"`
}

// Award represents something earned from a single activity (badge or level-up).
type Award struct {
	Type       string `json:"type"` // "badge" or "level_up"
	Badge      *Badge `json:"badge,omitempty"`
	Level      int    `json:"level,omitempty"`
	LevelTitle string `json:"level_title,omitempty"`
}

// Condition config structs for badge evaluation.

// CountCondition requires a threshold count of a specific entity type.
type CountCondition struct {
	Entity    string `json:"entity"`
	Threshold int    `json:"threshold"`
}

// StreakCondition requires maintaining a streak for a number of days.
type StreakCondition struct {
	Days int `json:"days"`
}

// ActionCondition matches a specific one-time action.
type ActionCondition struct {
	Action string `json:"action"`
}

// LevelThresholds defines the point thresholds for each level.
var LevelThresholds = []struct {
	Level  int
	Points int
	Title  string
}{
	{1, 0, "Newcomer"},
	{2, 100, "Contributor"},
	{3, 300, "Achiever"},
	{4, 600, "Career Builder"},
	{5, 1000, "Career Architect"},
	{6, 2000, "Career Legend"},
}

// ActionPoints maps activity actions to their point values.
var ActionPoints = map[string]int{
	"win_created":     10,
	"job_created":     15,
	"job_updated":     15,
	"cert_created":    10,
	"cert_passed":     25,
	"skill_created":   5,
	"evidence_linked": 10,
	"badge_earned":    50,
}

// LevelForPoints returns the level and title for a given point total.
func LevelForPoints(points int) (int, string) {
	level := 1
	title := "Newcomer"
	for _, t := range LevelThresholds {
		if points >= t.Points {
			level = t.Level
			title = t.Title
		}
	}
	return level, title
}

// NextLevelInfo returns the next level number and the points required to reach it.
func NextLevelInfo(currentLevel int) (int, int) {
	for _, t := range LevelThresholds {
		if t.Level > currentLevel {
			return t.Level, t.Points
		}
	}
	return currentLevel, 0 // max level
}
