package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Pipeline stages an application moves through. Order here is the canonical
// board order; accepted and rejected are terminal.
const (
	StatusWishlist  = "wishlist"
	StatusApplied   = "applied"
	StatusScreen    = "screen"
	StatusInterview = "interview"
	StatusOffer     = "offer"
	StatusAccepted  = "accepted"
	StatusRejected  = "rejected"
)

var validStatuses = []string{
	StatusWishlist, StatusApplied, StatusScreen, StatusInterview,
	StatusOffer, StatusAccepted, StatusRejected,
}

// Reused from the job domain's values so the two stay consistent.
var validWorkModes = []string{
	"onsite", "hybrid", "remote",
}

// repository is the subset of persistence the service needs, declared here so it
// can be mocked in tests.
type repository interface {
	List(ctx context.Context, userID uuid.UUID) ([]Application, error)
	GetByID(ctx context.Context, userID, applicationID uuid.UUID) (Application, error)
	Create(ctx context.Context, a *Application) error
	Update(ctx context.Context, a *Application) error
	Move(ctx context.Context, userID, applicationID uuid.UUID, status string, sortOrder int) error
	MaxSortOrder(ctx context.Context, userID uuid.UUID, status string) (int, error)
	Reorder(ctx context.Context, userID uuid.UUID, status string, orderedIDs []uuid.UUID) error
	Delete(ctx context.Context, userID, applicationID uuid.UUID) error
}

type Service struct {
	repo repository
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Application, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) GetByID(ctx context.Context, userID, applicationID uuid.UUID) (Application, error) {
	return s.repo.GetByID(ctx, userID, applicationID)
}

func (s *Service) Create(ctx context.Context, a *Application) error {
	if a.Status == "" {
		a.Status = StatusWishlist
	}
	if !isValidStatus(a.Status) {
		return fmt.Errorf("invalid status: %s", a.Status)
	}
	if a.WorkMode != nil && *a.WorkMode != "" && !isValidWorkMode(*a.WorkMode) {
		return fmt.Errorf("invalid work_mode: %s", *a.WorkMode)
	}
	a.ID = uuid.New()
	return s.repo.Create(ctx, a)
}

func (s *Service) Update(ctx context.Context, a *Application) error {
	if a.Status != "" && !isValidStatus(a.Status) {
		return fmt.Errorf("invalid status: %s", a.Status)
	}
	if a.WorkMode != nil && *a.WorkMode != "" && !isValidWorkMode(*a.WorkMode) {
		return fmt.Errorf("invalid work_mode: %s", *a.WorkMode)
	}
	return s.repo.Update(ctx, a)
}

// Move changes an application's stage and appends it to the end of the target
// column. Used when a drop carries no explicit ordering.
func (s *Service) Move(ctx context.Context, userID, applicationID uuid.UUID, status string) error {
	if !isValidStatus(status) {
		return fmt.Errorf("invalid status: %s", status)
	}
	maxOrder, err := s.repo.MaxSortOrder(ctx, userID, status)
	if err != nil {
		return err
	}
	return s.repo.Move(ctx, userID, applicationID, status, maxOrder+1)
}

// Reorder rewrites a whole column to the given order. The moved card must be
// part of orderedIDs when a drop crosses columns.
func (s *Service) Reorder(ctx context.Context, userID uuid.UUID, status string, orderedIDs []uuid.UUID) error {
	if !isValidStatus(status) {
		return fmt.Errorf("invalid status: %s", status)
	}
	return s.repo.Reorder(ctx, userID, status, orderedIDs)
}

func (s *Service) Delete(ctx context.Context, userID, applicationID uuid.UUID) error {
	return s.repo.Delete(ctx, userID, applicationID)
}

func isValidStatus(status string) bool {
	for _, v := range validStatuses {
		if v == status {
			return true
		}
	}
	return false
}

func isValidWorkMode(mode string) bool {
	for _, v := range validWorkModes {
		if v == mode {
			return true
		}
	}
	return false
}
