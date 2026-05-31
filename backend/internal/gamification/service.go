package gamification

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/trackmycareer/app/pkg/types"
)

// Service orchestrates gamification logic: points, streaks, badges, and activity logging.
type Service struct {
	repo *Repository
}

// NewService creates a new gamification service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// RecordActivity logs an action, awards points, updates the streak, and checks badge conditions.
// It returns any awards earned as a result (badges, level-ups).
func (s *Service) RecordActivity(ctx context.Context, userID uuid.UUID, action string, entityID *uuid.UUID) ([]Award, error) {
	var awards []Award

	// 1. Log the activity.
	if err := s.repo.LogActivity(ctx, userID, action, entityID); err != nil {
		return nil, fmt.Errorf("logging activity: %w", err)
	}

	// 2. Award points for the action, subject to a 24-hour cooldown per
	//    (user, action, entity) triple to prevent point farming.
	points := ActionPoints[action]
	if points > 0 {
		cooldownWindow := time.Now().UTC().Add(-24 * time.Hour)
		recentlyAwarded, err := s.repo.HasRecentActivity(ctx, userID, action, entityID, cooldownWindow)
		if err != nil {
			return nil, fmt.Errorf("checking recent activity: %w", err)
		}

		if !recentlyAwarded {
			prevPoints, err := s.repo.GetOrCreatePoints(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("fetching points before add: %w", err)
			}
			prevLevel := prevPoints.Level

			userPoints, err := s.repo.AddPoints(ctx, userID, points)
			if err != nil {
				return nil, fmt.Errorf("adding points: %w", err)
			}

			// Check for level-up.
			newLevel, title := LevelForPoints(userPoints.TotalPoints)
			if newLevel > prevLevel {
				awards = append(awards, Award{
					Type:       "level_up",
					Level:      newLevel,
					LevelTitle: title,
				})
			}
		}
	}

	// 3. Update the streak.
	if err := s.updateStreak(ctx, userID); err != nil {
		return nil, fmt.Errorf("updating streak: %w", err)
	}

	// 4. Check badge conditions.
	badgeAwards, err := s.checkBadges(ctx, userID, action)
	if err != nil {
		return nil, fmt.Errorf("checking badges: %w", err)
	}
	awards = append(awards, badgeAwards...)

	// 5. Award bonus points for each badge earned.
	//    Badges cannot trigger further badges, so this is not recursive.
	badgeBonusPoints := ActionPoints["badge_earned"]
	for range badgeAwards {
		if badgeBonusPoints > 0 {
			if _, err := s.repo.AddPoints(ctx, userID, badgeBonusPoints); err != nil {
				return nil, fmt.Errorf("adding badge bonus points: %w", err)
			}
		}
	}

	return awards, nil
}

// RecalculationResult summarises the outcome of a bulk badge recalculation.
type RecalculationResult struct {
	UsersProcessed int `json:"users_processed"`
	BadgesAwarded  int `json:"badges_awarded"`
}

// RecalculateAllBadges re-evaluates badge conditions for every user and awards
// any badges whose conditions are now satisfied. Safe to call repeatedly because
// badge awards are idempotent.
func (s *Service) RecalculateAllBadges(ctx context.Context) (*RecalculationResult, error) {
	userIDs, err := s.repo.ListAllUserIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing users for recalculation: %w", err)
	}

	result := &RecalculationResult{UsersProcessed: len(userIDs)}
	for _, uid := range userIDs {
		awards, err := s.recalculateBadgesForUser(ctx, uid)
		if err != nil {
			slog.Warn("badge recalculation failed for user", "error", err.Error(), "user_id", uid.String())
			continue
		}
		result.BadgesAwarded += len(awards)
	}

	slog.Info("badge recalculation complete", "users_processed", result.UsersProcessed, "badges_awarded", result.BadgesAwarded)
	return result, nil
}

// GetProgress returns the aggregated gamification overview for a user.
func (s *Service) GetProgress(ctx context.Context, userID uuid.UUID) (*Progress, error) {
	points, err := s.repo.GetOrCreatePoints(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching points: %w", err)
	}

	streak, err := s.repo.GetOrCreateStreak(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching streak: %w", err)
	}

	badgeCount, err := s.repo.CountUserBadges(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("counting badges: %w", err)
	}

	level, title := LevelForPoints(points.TotalPoints)
	nextLevel, nextLevelAt := NextLevelInfo(level)

	return &Progress{
		TotalPoints:   points.TotalPoints,
		Level:         level,
		LevelTitle:    title,
		NextLevel:     nextLevel,
		NextLevelAt:   nextLevelAt,
		BadgesEarned:  badgeCount,
		CurrentStreak: streak.CurrentStreak,
		LongestStreak: streak.LongestStreak,
	}, nil
}

// GetStreak returns the user's current streak details.
func (s *Service) GetStreak(ctx context.Context, userID uuid.UUID) (*UserStreak, error) {
	return s.repo.GetOrCreateStreak(ctx, userID)
}

// ListBadgesForUser returns all badge definitions with per-user earned status.
func (s *Service) ListBadgesForUser(ctx context.Context, userID uuid.UUID) ([]Badge, error) {
	return s.repo.ListBadgesWithEarnedStatus(ctx, userID)
}

// GetHeatmap returns daily activity counts for the specified number of past days.
func (s *Service) GetHeatmap(ctx context.Context, userID uuid.UUID, days int) ([]HeatmapEntry, error) {
	if days <= 0 {
		days = 365
	}
	return s.repo.GetHeatmap(ctx, userID, days)
}

// updateStreak recalculates the user's streak based on today's date and the last active date.
func (s *Service) updateStreak(ctx context.Context, userID uuid.UUID) error {
	streak, err := s.repo.GetOrCreateStreak(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	today := types.NewDate(now.Truncate(24 * time.Hour))

	if streak.LastActiveOn != nil && streak.LastActiveOn.Time.Equal(today.Time) {
		return nil
	}

	yesterday := types.NewDate(now.AddDate(0, 0, -1).Truncate(24 * time.Hour))

	if streak.LastActiveOn != nil && streak.LastActiveOn.Time.Equal(yesterday.Time) {
		streak.CurrentStreak++
	} else {
		streak.CurrentStreak = 1
	}

	if streak.CurrentStreak > streak.LongestStreak {
		streak.LongestStreak = streak.CurrentStreak
	}

	streak.LastActiveOn = &today

	return s.repo.UpdateStreak(ctx, streak)
}
