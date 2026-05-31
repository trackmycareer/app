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

// badgeActionToEntity maps badge action condition keys back to their entity
// type for recalculation (where there is no "current action" to match against).
var badgeActionToEntity = map[string]string{
	"first_win":           "win",
	"first_job":           "job",
	"first_cert":          "certification",
	"first_cert_passed":   "cert_passed",
	"first_skill":         "skill",
	"first_evidence_link": "evidence",
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

// recalculateBadgesForUser evaluates all unearned badges for a user without
// requiring a triggering action. Used by the admin recalculation endpoint.
func (s *Service) recalculateBadgesForUser(ctx context.Context, userID uuid.UUID) ([]Award, error) {
	badges, err := s.repo.ListBadgesWithEarnedStatus(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing badges for recalculation: %w", err)
	}

	var awards []Award
	for i := range badges {
		b := &badges[i]
		if b.Earned {
			continue
		}

		satisfied, err := s.evaluateConditionForRecalc(ctx, userID, b)
		if err != nil {
			slog.Warn("badge recalculation evaluation failed", "error", err.Error(), "badge", b.Name, "user_id", userID.String())
			continue
		}
		if !satisfied {
			continue
		}

		if err := s.repo.AwardBadge(ctx, userID, b.ID); err != nil {
			slog.Error("badge recalculation award failed", "error", err.Error(), "badge", b.Name, "user_id", userID.String())
			continue
		}

		awards = append(awards, Award{
			Type:  "badge",
			Badge: b,
		})
	}

	return awards, nil
}

// evaluateConditionForRecalc checks whether a badge condition is satisfied
// without a triggering action. For action conditions, it checks entity count >= 1
// rather than requiring a matching current action with count == 1.
func (s *Service) evaluateConditionForRecalc(ctx context.Context, userID uuid.UUID, b *Badge) (bool, error) {
	switch b.ConditionType {
	case "action":
		var cond ActionCondition
		if err := json.Unmarshal(b.ConditionConfig, &cond); err != nil {
			return false, fmt.Errorf("unmarshalling action condition: %w", err)
		}
		entity, ok := badgeActionToEntity[cond.Action]
		if !ok {
			return false, nil
		}
		count, err := s.repo.GetEntityCount(ctx, userID, entity)
		if err != nil {
			return false, err
		}
		return count >= 1, nil
	case "count":
		return s.evaluateCountCondition(ctx, userID, b)
	case "streak":
		return s.evaluateStreakCondition(ctx, userID, b)
	default:
		return false, fmt.Errorf("unknown condition type: %s", b.ConditionType)
	}
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
