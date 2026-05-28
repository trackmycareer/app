package certification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

var validStatuses = []string{
	"planning", "studying", "scheduled", "passed", "expired",
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, params ListParams) ([]Certification, int, error) {
	if params.Status != "" && !isValidStatus(params.Status) {
		return nil, 0, fmt.Errorf("invalid status: %s", params.Status)
	}
	return s.repo.List(ctx, userID, params)
}

func (s *Service) GetByID(ctx context.Context, userID, certID uuid.UUID) (Certification, error) {
	return s.repo.GetByID(ctx, userID, certID)
}

func (s *Service) Create(ctx context.Context, c *Certification) error {
	if c.Status == "" {
		c.Status = "planning"
	}
	if !isValidStatus(c.Status) {
		return fmt.Errorf("invalid status: %s", c.Status)
	}
	if c.Currency == "" {
		c.Currency = "GBP"
	}
	c.ID = uuid.New()
	return s.repo.Create(ctx, c)
}

func (s *Service) Update(ctx context.Context, c *Certification) error {
	if c.Status != "" && !isValidStatus(c.Status) {
		return fmt.Errorf("invalid status: %s", c.Status)
	}
	if c.Currency == "" {
		c.Currency = "GBP"
	}
	return s.repo.Update(ctx, c)
}

func (s *Service) UpdateStatus(ctx context.Context, userID, certID uuid.UUID, status string) error {
	if !isValidStatus(status) {
		return fmt.Errorf("invalid status: %s", status)
	}
	return s.repo.UpdateStatus(ctx, userID, certID, status)
}

func (s *Service) Delete(ctx context.Context, userID, certID uuid.UUID) error {
	return s.repo.Delete(ctx, userID, certID)
}

func isValidStatus(status string) bool {
	for _, v := range validStatuses {
		if v == status {
			return true
		}
	}
	return false
}
