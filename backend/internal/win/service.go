package win

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/types"
)

var validCategories = []string{
	"shipped_feature", "positive_feedback", "process_improvement",
	"cost_saving", "leadership_moment", "general",
	"project", "publication", "membership",
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, params ListParams) ([]Win, int, error) {
	if params.Category != "" && !isValidCategory(params.Category) {
		return nil, 0, fmt.Errorf("invalid category: %s", params.Category)
	}
	return s.repo.List(ctx, userID, params)
}

func (s *Service) GetByID(ctx context.Context, userID, winID uuid.UUID) (Win, error) {
	return s.repo.GetByID(ctx, userID, winID)
}

func (s *Service) Create(ctx context.Context, w *Win, tagIDs []uuid.UUID) error {
	if w.OccurredOn.IsZero() {
		w.OccurredOn = types.Today()
	}
	if w.Category == "" {
		w.Category = "general"
	}
	if !isValidCategory(w.Category) {
		return fmt.Errorf("invalid category: %s", w.Category)
	}
	w.ID = uuid.New()
	return s.repo.Create(ctx, w, tagIDs)
}

func (s *Service) Update(ctx context.Context, w *Win, tagIDs []uuid.UUID) error {
	if w.Category != "" && !isValidCategory(w.Category) {
		return fmt.Errorf("invalid category: %s", w.Category)
	}
	return s.repo.Update(ctx, w, tagIDs)
}

func (s *Service) Delete(ctx context.Context, userID, winID uuid.UUID) error {
	return s.repo.Delete(ctx, userID, winID)
}

func isValidCategory(cat string) bool {
	for _, v := range validCategories {
		if v == cat {
			return true
		}
	}
	return false
}
