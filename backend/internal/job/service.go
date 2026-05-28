package job

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

var validEmploymentTypes = []string{
	"full_time", "part_time", "contract", "freelance", "internship",
}

var validTransitionTypes = []string{
	"promotion", "lateral_move", "company_change", "first_role",
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Job, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) GetByID(ctx context.Context, userID, jobID uuid.UUID) (Job, error) {
	return s.repo.GetByID(ctx, userID, jobID)
}

func (s *Service) Create(ctx context.Context, j *Job) error {
	if j.EmploymentType == "" {
		j.EmploymentType = "full_time"
	}
	if !isValidEmploymentType(j.EmploymentType) {
		return fmt.Errorf("invalid employment_type: %s", j.EmploymentType)
	}
	if j.TransitionType != nil && !isValidTransitionType(*j.TransitionType) {
		return fmt.Errorf("invalid transition_type: %s", *j.TransitionType)
	}
	j.ID = uuid.New()
	return s.repo.Create(ctx, j)
}

func (s *Service) Update(ctx context.Context, j *Job) error {
	if j.EmploymentType != "" && !isValidEmploymentType(j.EmploymentType) {
		return fmt.Errorf("invalid employment_type: %s", j.EmploymentType)
	}
	if j.TransitionType != nil && !isValidTransitionType(*j.TransitionType) {
		return fmt.Errorf("invalid transition_type: %s", *j.TransitionType)
	}
	return s.repo.Update(ctx, j)
}

func (s *Service) Delete(ctx context.Context, userID, jobID uuid.UUID) error {
	return s.repo.Delete(ctx, userID, jobID)
}

func isValidEmploymentType(t string) bool {
	for _, v := range validEmploymentTypes {
		if v == t {
			return true
		}
	}
	return false
}

func isValidTransitionType(t string) bool {
	for _, v := range validTransitionTypes {
		if v == t {
			return true
		}
	}
	return false
}
