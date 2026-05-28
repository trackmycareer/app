package skill

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

var validEvidenceTypes = []string{"win", "certification", "job"}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, params ListParams) ([]Skill, int, error) {
	return s.repo.List(ctx, userID, params)
}

func (s *Service) GetByID(ctx context.Context, userID, skillID uuid.UUID) (Skill, error) {
	return s.repo.GetByID(ctx, userID, skillID)
}

func (s *Service) Create(ctx context.Context, sk *Skill) error {
	if sk.Proficiency < 1 || sk.Proficiency > 4 {
		return fmt.Errorf("invalid proficiency: must be between 1 and 4")
	}
	sk.ID = uuid.New()
	return s.repo.Create(ctx, sk)
}

func (s *Service) Update(ctx context.Context, sk *Skill) error {
	if sk.Proficiency < 1 || sk.Proficiency > 4 {
		return fmt.Errorf("invalid proficiency: must be between 1 and 4")
	}
	return s.repo.Update(ctx, sk)
}

func (s *Service) Delete(ctx context.Context, userID, skillID uuid.UUID) error {
	return s.repo.Delete(ctx, userID, skillID)
}

func (s *Service) AddEvidence(ctx context.Context, userID, skillID uuid.UUID, evidenceType string, evidenceID uuid.UUID) (SkillEvidence, error) {
	if !isValidEvidenceType(evidenceType) {
		return SkillEvidence{}, fmt.Errorf("invalid evidence_type: must be one of win, certification, job")
	}
	return s.repo.AddEvidence(ctx, userID, skillID, evidenceType, evidenceID)
}

func (s *Service) RemoveEvidence(ctx context.Context, userID, skillID, evidenceID uuid.UUID) error {
	return s.repo.RemoveEvidence(ctx, userID, skillID, evidenceID)
}

func isValidEvidenceType(t string) bool {
	for _, v := range validEvidenceTypes {
		if v == t {
			return true
		}
	}
	return false
}
