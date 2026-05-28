package gamification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// actionToBadgeAction maps activity actions to the badge action-condition keys,
// but only when the entity count is exactly 1 (i.e. first occurrence).
var actionToBadgeAction = map[string]struct {
	badgeAction string
	countEntity string
}{
	"win_created":     {badgeAction: "first_win", countEntity: "win"},
	"job_created":     {badgeAction: "first_job", countEntity: "job"},
	"cert_created":    {badgeAction: "first_cert", countEntity: "certification"},
	"cert_passed":     {badgeAction: "first_cert_passed", countEntity: "cert_passed"},
	"skill_created":   {badgeAction: "first_skill", countEntity: "skill"},
	"evidence_linked": {badgeAction: "first_evidence_link", countEntity: "evidence"},
}

// checkBadges evaluates all unearned badges against the user's current state
// and awards any whose conditions are now satisfied.
func (s *Service) checkBadges(ctx context.Context, userID uuid.UUID, action string) ([]Award, error) {
	// Fetch all badges with earned status for this user.
	badges, err := s.repo.ListBadgesWithEarnedStatus(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing badges for check: %w", err)
	}

	var awards []Award
	for i := range badges {
		b := &badges[i]
		if b.Earned {
			continue
		}

		satisfied, err := s.evaluateCondition(ctx, userID, b, action)
		if err != nil {
			slog.Warn("badge condition evaluation failed", "error", err.Error(), "badge", b.Name, "user_id", userID.String())
			continue
		}
		if !satisfied {
			continue
		}

		if err := s.repo.AwardBadge(ctx, userID, b.ID); err != nil {
			slog.Error("badge award failed", "error", err.Error(), "badge", b.Name, "user_id", userID.String())
			continue
		}

		awards = append(awards, Award{
			Type:  "badge",
			Badge: b,
		})
	}

	return awards, nil
}

// evaluateCondition checks whether a badge's condition is satisfied.
func (s *Service) evaluateCondition(ctx context.Context, userID uuid.UUID, b *Badge, action string) (bool, error) {
	switch b.ConditionType {
	case "action":
		return s.evaluateActionCondition(ctx, userID, b, action)
	case "count":
		return s.evaluateCountCondition(ctx, userID, b)
	case "streak":
		return s.evaluateStreakCondition(ctx, userID, b)
	default:
		return false, fmt.Errorf("unknown condition type: %s", b.ConditionType)
	}
}

// evaluateActionCondition checks if the current action matches a first-occurrence badge.
func (s *Service) evaluateActionCondition(ctx context.Context, userID uuid.UUID, b *Badge, action string) (bool, error) {
	var cond ActionCondition
	if err := json.Unmarshal(b.ConditionConfig, &cond); err != nil {
		return false, fmt.Errorf("unmarshalling action condition: %w", err)
	}

	mapping, ok := actionToBadgeAction[action]
	if !ok {
		return false, nil
	}

	if mapping.badgeAction != cond.Action {
		return false, nil
	}

	// Verify this is genuinely the first occurrence.
	count, err := s.repo.GetEntityCount(ctx, userID, mapping.countEntity)
	if err != nil {
		return false, err
	}

	return count == 1, nil
}

// evaluateCountCondition checks if the user's entity count meets the badge threshold.
func (s *Service) evaluateCountCondition(ctx context.Context, userID uuid.UUID, b *Badge) (bool, error) {
	var cond CountCondition
	if err := json.Unmarshal(b.ConditionConfig, &cond); err != nil {
		return false, fmt.Errorf("unmarshalling count condition: %w", err)
	}

	count, err := s.repo.GetEntityCount(ctx, userID, cond.Entity)
	if err != nil {
		return false, err
	}

	return count >= cond.Threshold, nil
}

// evaluateStreakCondition checks if the user's current streak meets the required days.
func (s *Service) evaluateStreakCondition(ctx context.Context, userID uuid.UUID, b *Badge) (bool, error) {
	var cond StreakCondition
	if err := json.Unmarshal(b.ConditionConfig, &cond); err != nil {
		return false, fmt.Errorf("unmarshalling streak condition: %w", err)
	}

	streak, err := s.repo.GetOrCreateStreak(ctx, userID)
	if err != nil {
		return false, err
	}

	return streak.CurrentStreak >= cond.Days, nil
}
