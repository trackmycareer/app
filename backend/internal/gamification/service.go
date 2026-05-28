package gamification

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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

	// 2. Award points for the action.
	points := ActionPoints[action]
	if points > 0 {
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

	today := time.Now().UTC().Format("2006-01-02")

	// If already active today, nothing to do.
	if streak.LastActiveOn != nil && *streak.LastActiveOn == today {
		return nil
	}

	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")

	if streak.LastActiveOn != nil && *streak.LastActiveOn == yesterday {
		// Consecutive day: increment streak.
		streak.CurrentStreak++
	} else {
		// Gap or first activity: reset to 1.
		streak.CurrentStreak = 1
	}

	if streak.CurrentStreak > streak.LongestStreak {
		streak.LongestStreak = streak.CurrentStreak
	}

	streak.LastActiveOn = &today

	return s.repo.UpdateStreak(ctx, streak)
}
