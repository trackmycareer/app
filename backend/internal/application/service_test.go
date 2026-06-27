package application

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// mockRepo is a configurable test double for the repository interface. It
// records the values passed to the mutating methods so tests can assert on them.
type mockRepo struct {
	maxSortOrder int
	maxErr       error
	createErr    error
	updateErr    error
	moveErr      error
	reorderErr   error

	created       *Application
	movedStatus   string
	movedOrder    int
	reorderStatus string
	reorderIDs    []uuid.UUID
}

func (m *mockRepo) List(_ context.Context, _ uuid.UUID) ([]Application, error) {
	return nil, nil
}

func (m *mockRepo) GetByID(_ context.Context, _, _ uuid.UUID) (Application, error) {
	return Application{}, nil
}

func (m *mockRepo) Create(_ context.Context, a *Application) error {
	m.created = a
	return m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ *Application) error {
	return m.updateErr
}

func (m *mockRepo) Move(_ context.Context, _, _ uuid.UUID, status string, sortOrder int) error {
	m.movedStatus = status
	m.movedOrder = sortOrder
	return m.moveErr
}

func (m *mockRepo) MaxSortOrder(_ context.Context, _ uuid.UUID, _ string) (int, error) {
	return m.maxSortOrder, m.maxErr
}

func (m *mockRepo) Reorder(_ context.Context, _ uuid.UUID, status string, orderedIDs []uuid.UUID) error {
	m.reorderStatus = status
	m.reorderIDs = orderedIDs
	return m.reorderErr
}

func (m *mockRepo) Delete(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func strptr(s string) *string { return &s }

func TestServiceCreate(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name       string
		input      Application
		wantErr    bool
		wantErrSub string
		wantStatus string
	}{
		{
			name:       "empty status defaults to wishlist",
			input:      Application{UserID: userID, Company: "Acme", Title: "Engineer"},
			wantStatus: StatusWishlist,
		},
		{
			name:       "valid explicit status is kept",
			input:      Application{UserID: userID, Company: "Acme", Title: "Engineer", Status: StatusApplied},
			wantStatus: StatusApplied,
		},
		{
			name:       "invalid status is rejected",
			input:      Application{UserID: userID, Company: "Acme", Title: "Engineer", Status: "bogus"},
			wantErr:    true,
			wantErrSub: "invalid status",
		},
		{
			name:       "valid work mode is accepted",
			input:      Application{UserID: userID, Company: "Acme", Title: "Engineer", WorkMode: strptr("remote")},
			wantStatus: StatusWishlist,
		},
		{
			name:       "invalid work mode is rejected",
			input:      Application{UserID: userID, Company: "Acme", Title: "Engineer", WorkMode: strptr("space_station")},
			wantErr:    true,
			wantErrSub: "invalid work_mode",
		},
		{
			name:       "empty work mode pointer is ignored",
			input:      Application{UserID: userID, Company: "Acme", Title: "Engineer", WorkMode: strptr("")},
			wantStatus: StatusWishlist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			svc := NewService(repo)

			a := tt.input
			err := svc.Create(ctx, &a)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrSub != "" && !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantErrSub)
				}
				if repo.created != nil {
					t.Fatalf("repo.Create should not be called on validation failure")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", a.Status, tt.wantStatus)
			}
			if a.ID == uuid.Nil {
				t.Errorf("expected an ID to be assigned")
			}
		})
	}
}

func TestServiceUpdate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		input      Application
		wantErr    bool
		wantErrSub string
	}{
		{
			name:  "empty status skips validation",
			input: Application{Company: "Acme", Title: "Engineer"},
		},
		{
			name:    "invalid status is rejected",
			input:   Application{Status: "bogus"},
			wantErr: true, wantErrSub: "invalid status",
		},
		{
			name:    "invalid work mode is rejected",
			input:   Application{Status: StatusOffer, WorkMode: strptr("teleport")},
			wantErr: true, wantErrSub: "invalid work_mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&mockRepo{})
			a := tt.input
			err := svc.Update(ctx, &a)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrSub != "" && !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestServiceMove(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	appID := uuid.New()

	t.Run("invalid status is rejected before touching repo", func(t *testing.T) {
		repo := &mockRepo{}
		svc := NewService(repo)
		err := svc.Move(ctx, userID, appID, "bogus")
		if err == nil || !strings.Contains(err.Error(), "invalid status") {
			t.Fatalf("error = %v, want invalid status", err)
		}
		if repo.movedStatus != "" {
			t.Fatalf("repo.Move should not be called on invalid status")
		}
	})

	t.Run("appends to end of column using max sort order", func(t *testing.T) {
		repo := &mockRepo{maxSortOrder: 4}
		svc := NewService(repo)
		if err := svc.Move(ctx, userID, appID, StatusInterview); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.movedStatus != StatusInterview {
			t.Errorf("moved status = %q, want %q", repo.movedStatus, StatusInterview)
		}
		if repo.movedOrder != 5 {
			t.Errorf("moved order = %d, want 5 (max+1)", repo.movedOrder)
		}
	})

	t.Run("empty column appends at position zero", func(t *testing.T) {
		repo := &mockRepo{maxSortOrder: -1}
		svc := NewService(repo)
		if err := svc.Move(ctx, userID, appID, StatusApplied); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.movedOrder != 0 {
			t.Errorf("moved order = %d, want 0", repo.movedOrder)
		}
	})
}

func TestServiceReorder(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	t.Run("invalid status is rejected before touching repo", func(t *testing.T) {
		repo := &mockRepo{}
		svc := NewService(repo)
		err := svc.Reorder(ctx, userID, "bogus", ids)
		if err == nil || !strings.Contains(err.Error(), "invalid status") {
			t.Fatalf("error = %v, want invalid status", err)
		}
		if repo.reorderIDs != nil {
			t.Fatalf("repo.Reorder should not be called on invalid status")
		}
	})

	t.Run("forwards status and order to repo", func(t *testing.T) {
		repo := &mockRepo{}
		svc := NewService(repo)
		if err := svc.Reorder(ctx, userID, StatusOffer, ids); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.reorderStatus != StatusOffer {
			t.Errorf("reorder status = %q, want %q", repo.reorderStatus, StatusOffer)
		}
		if len(repo.reorderIDs) != len(ids) {
			t.Fatalf("reorder ids len = %d, want %d", len(repo.reorderIDs), len(ids))
		}
		for i := range ids {
			if repo.reorderIDs[i] != ids[i] {
				t.Errorf("reorder id[%d] = %s, want %s", i, repo.reorderIDs[i], ids[i])
			}
		}
	})
}
