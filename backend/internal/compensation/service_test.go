package compensation

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/mfa/crypto"
)

// mockRepo is an in-memory stand-in for the pgx repository. It mimics the real
// store by persisting only the ciphertext: Amounts is zeroed on write, so reads
// can only recover the values by decrypting, which is what we want to exercise.
type mockRepo struct {
	store        map[uuid.UUID]Compensation
	createErr    error
	updateErr    error
	createCalled bool
	updateCalled bool
}

func newMockRepo() *mockRepo {
	return &mockRepo{store: make(map[uuid.UUID]Compensation)}
}

func (m *mockRepo) List(_ context.Context, userID uuid.UUID) ([]Compensation, error) {
	var out []Compensation
	for _, c := range m.store {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *mockRepo) GetByID(_ context.Context, userID, id uuid.UUID) (Compensation, error) {
	c, ok := m.store[id]
	if !ok || c.UserID != userID {
		return Compensation{}, fmt.Errorf("compensation not found")
	}
	return c, nil
}

func (m *mockRepo) Create(_ context.Context, c *Compensation) error {
	m.createCalled = true
	if m.createErr != nil {
		return m.createErr
	}
	stored := *c
	stored.Amounts = Amounts{}
	m.store[c.ID] = stored
	return nil
}

func (m *mockRepo) Update(_ context.Context, c *Compensation) error {
	m.updateCalled = true
	if m.updateErr != nil {
		return m.updateErr
	}
	stored := *c
	stored.Amounts = Amounts{}
	m.store[c.ID] = stored
	return nil
}

func (m *mockRepo) Delete(_ context.Context, _, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func testEncryptor(t *testing.T) *crypto.Encryptor {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generating key: %v", err)
	}
	enc, err := crypto.NewEncryptor(key)
	if err != nil {
		t.Fatalf("creating encryptor: %v", err)
	}
	return enc
}

func strptr(s string) *string { return &s }

func TestServiceCreateValidation(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name         string
		input        Compensation
		wantErrSub   string
		wantCurrency string
		wantBasis    string
		wantCreate   bool
	}{
		{
			name:         "valid with explicit values",
			input:        Compensation{UserID: userID, JobID: uuid.New(), Currency: "USD", PayBasis: "annual", Amounts: Amounts{Base: 10000000}},
			wantCurrency: "USD",
			wantBasis:    "annual",
			wantCreate:   true,
		},
		{
			name:         "currency defaults to GBP",
			input:        Compensation{UserID: userID, JobID: uuid.New(), Amounts: Amounts{Base: 5000000}},
			wantCurrency: "GBP",
			wantBasis:    "annual",
			wantCreate:   true,
		},
		{
			name:         "currency is uppercased",
			input:        Compensation{UserID: userID, JobID: uuid.New(), Currency: "eur", PayBasis: "monthly", Amounts: Amounts{Base: 400000}},
			wantCurrency: "EUR",
			wantBasis:    "monthly",
			wantCreate:   true,
		},
		{
			name:       "invalid currency length",
			input:      Compensation{UserID: userID, JobID: uuid.New(), Currency: "US", Amounts: Amounts{Base: 1}},
			wantErrSub: "currency must be",
		},
		{
			name:       "invalid currency non-alpha",
			input:      Compensation{UserID: userID, JobID: uuid.New(), Currency: "US1", Amounts: Amounts{Base: 1}},
			wantErrSub: "currency must be",
		},
		{
			name:       "invalid pay basis",
			input:      Compensation{UserID: userID, JobID: uuid.New(), PayBasis: "fortnightly", Amounts: Amounts{Base: 1}},
			wantErrSub: "invalid pay_basis",
		},
		{
			name:       "negative base rejected",
			input:      Compensation{UserID: userID, JobID: uuid.New(), Amounts: Amounts{Base: -1}},
			wantErrSub: "amounts must be non-negative",
		},
		{
			name:       "negative bonus rejected",
			input:      Compensation{UserID: userID, JobID: uuid.New(), Amounts: Amounts{Bonus: -500}},
			wantErrSub: "amounts must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepo()
			svc := NewService(repo, testEncryptor(t))

			c := tt.input
			err := svc.Create(ctx, &c)

			if tt.wantErrSub != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Fatalf("Create() error = %v, want substring %q", err, tt.wantErrSub)
				}
				if repo.createCalled {
					t.Errorf("repo.Create should not be called when validation fails")
				}
				return
			}

			if err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}
			if !repo.createCalled {
				t.Errorf("repo.Create was not called")
			}
			if c.Currency != tt.wantCurrency {
				t.Errorf("currency = %q, want %q", c.Currency, tt.wantCurrency)
			}
			if c.PayBasis != tt.wantBasis {
				t.Errorf("pay_basis = %q, want %q", c.PayBasis, tt.wantBasis)
			}
			if c.ID == uuid.Nil {
				t.Errorf("ID was not assigned")
			}
			if len(c.EncryptedData) == 0 || len(c.Nonce) == 0 {
				t.Errorf("expected encrypted_data and nonce to be populated")
			}
		})
	}
}

func TestNoteNormalisation(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := newMockRepo()
	svc := NewService(repo, testEncryptor(t))

	c := Compensation{UserID: userID, JobID: uuid.New(), Amounts: Amounts{Base: 100, Note: strptr("   ")}}
	if err := svc.Create(ctx, &c); err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if c.Amounts.Note != nil {
		t.Errorf("blank note should be normalised to nil, got %q", *c.Amounts.Note)
	}

	long := strings.Repeat("x", maxNoteLen+1)
	c2 := Compensation{UserID: userID, JobID: uuid.New(), Amounts: Amounts{Base: 100, Note: strptr(long)}}
	if err := svc.Create(ctx, &c2); err == nil || !strings.Contains(err.Error(), "note must be at most") {
		t.Errorf("expected note length error, got %v", err)
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := newMockRepo()
	svc := NewService(repo, testEncryptor(t))

	want := Amounts{Base: 9500000, Bonus: 1500000, Equity: 4000000, Other: 250000, Note: strptr("annual review uplift")}
	c := Compensation{UserID: userID, JobID: uuid.New(), Currency: "GBP", PayBasis: "annual", Amounts: want}
	if err := svc.Create(ctx, &c); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// The stored row must not contain the plaintext amounts.
	stored := repo.store[c.ID]
	if stored.Amounts != (Amounts{}) {
		t.Errorf("stored row unexpectedly retained plaintext amounts")
	}
	plaintextJSON, _ := json.Marshal(want)
	if bytes.Equal(stored.EncryptedData, plaintextJSON) {
		t.Errorf("encrypted_data equals plaintext JSON; data is not encrypted")
	}
	if bytes.Contains(stored.EncryptedData, []byte(`"base"`)) {
		t.Errorf("encrypted_data contains plaintext JSON keys")
	}

	// Reading back via the service must decrypt to the exact input.
	got, err := svc.GetByID(ctx, userID, c.ID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}
	assertAmountsEqual(t, got.Amounts, want)
	if len(got.EncryptedData) != 0 || len(got.Nonce) != 0 {
		t.Errorf("decrypted entry should not carry ciphertext back to the caller")
	}

	list, err := svc.List(ctx, userID)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("List() returned %d entries, want 1", len(list))
	}
	assertAmountsEqual(t, list[0].Amounts, want)
}

func TestNonceIsUniquePerWrite(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := newMockRepo()
	svc := NewService(repo, testEncryptor(t))

	amounts := Amounts{Base: 5000000}
	a := Compensation{UserID: userID, JobID: uuid.New(), Amounts: amounts}
	b := Compensation{UserID: userID, JobID: uuid.New(), Amounts: amounts}
	if err := svc.Create(ctx, &a); err != nil {
		t.Fatalf("Create(a) error: %v", err)
	}
	if err := svc.Create(ctx, &b); err != nil {
		t.Fatalf("Create(b) error: %v", err)
	}
	if bytes.Equal(a.Nonce, b.Nonce) {
		t.Errorf("two encryptions of identical amounts produced the same nonce")
	}
	if bytes.Equal(a.EncryptedData, b.EncryptedData) {
		t.Errorf("two encryptions of identical amounts produced identical ciphertext")
	}
}

func TestUpdateReEncrypts(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := newMockRepo()
	svc := NewService(repo, testEncryptor(t))

	c := Compensation{UserID: userID, JobID: uuid.New(), Amounts: Amounts{Base: 5000000}}
	if err := svc.Create(ctx, &c); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	c.Amounts = Amounts{Base: 6000000, Bonus: 1000000}
	if err := svc.Update(ctx, &c); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, err := svc.GetByID(ctx, userID, c.ID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}
	assertAmountsEqual(t, got.Amounts, Amounts{Base: 6000000, Bonus: 1000000})
}

func assertAmountsEqual(t *testing.T, got, want Amounts) {
	t.Helper()
	if got.Base != want.Base || got.Bonus != want.Bonus || got.Equity != want.Equity || got.Other != want.Other {
		t.Errorf("amounts = %+v, want %+v", got, want)
	}
	switch {
	case got.Note == nil && want.Note == nil:
	case got.Note == nil || want.Note == nil:
		t.Errorf("note = %v, want %v", got.Note, want.Note)
	case *got.Note != *want.Note:
		t.Errorf("note = %q, want %q", *got.Note, *want.Note)
	}
}
