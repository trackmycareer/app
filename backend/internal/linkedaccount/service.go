package linkedaccount

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID) ([]LinkedAccount, error) {
	accounts, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if accounts == nil {
		accounts = []LinkedAccount{}
	}
	return accounts, nil
}

func (s *Service) ListPublicByUser(ctx context.Context, userID uuid.UUID) ([]PublicLinkedAccount, error) {
	accounts, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var public []PublicLinkedAccount
	for _, a := range accounts {
		if a.Verified {
			public = append(public, PublicLinkedAccount{
				Provider:   a.Provider,
				ProfileURL: a.ProfileURL,
				Verified:   a.Verified,
				VerifiedAt: a.VerifiedAt,
			})
		}
	}
	if public == nil {
		public = []PublicLinkedAccount{}
	}
	return public, nil
}

func (s *Service) LinkOAuth(ctx context.Context, userID uuid.UUID, provider, providerID, profileURL string) (*LinkedAccount, error) {
	existing, err := s.repo.GetByProviderID(ctx, provider, providerID)
	if err == nil && existing.UserID != userID {
		return nil, fmt.Errorf("this %s account is already linked to another user", provider)
	}

	now := time.Now()
	la := &LinkedAccount{
		ID:         uuid.New(),
		UserID:     userID,
		Provider:   provider,
		ProviderID: &providerID,
		ProfileURL: profileURL,
		Verified:   true,
		VerifiedAt: &now,
		CreatedAt:  now,
	}

	if err := s.repo.Upsert(ctx, la); err != nil {
		return nil, err
	}
	return la, nil
}

func (s *Service) AddWebsite(ctx context.Context, userID uuid.UUID, rawURL string) (*LinkedAccount, error) {
	domain, err := ExtractDomain(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if err := ValidatePublicDomain(domain); err != nil {
		return nil, fmt.Errorf("invalid domain: %w", err)
	}

	if err := ValidateDomainResolution(domain); err != nil {
		return nil, fmt.Errorf("domain validation failed: %w", err)
	}

	token, err := generateVerifyToken()
	if err != nil {
		return nil, fmt.Errorf("generating verification token: %w", err)
	}

	fullToken := "trackmy-career-verify=" + token
	profileURL := "https://" + domain

	la := &LinkedAccount{
		ID:          uuid.New(),
		UserID:      userID,
		Provider:    ProviderWebsite,
		ProfileURL:  profileURL,
		Verified:    false,
		VerifyToken: &fullToken,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Upsert(ctx, la); err != nil {
		return nil, err
	}
	return la, nil
}

func (s *Service) VerifyWebsite(ctx context.Context, userID uuid.UUID) (*LinkedAccount, error) {
	la, err := s.repo.GetByUserAndProvider(ctx, userID, ProviderWebsite)
	if err != nil {
		return nil, fmt.Errorf("no website configured")
	}

	if la.VerifyToken == nil {
		return nil, fmt.Errorf("no verification token found")
	}

	domain, err := ExtractDomain(la.ProfileURL)
	if err != nil {
		return nil, fmt.Errorf("invalid stored URL: %w", err)
	}

	found, err := VerifyDNSTXT(domain, *la.VerifyToken)
	if err != nil {
		return nil, fmt.Errorf("DNS lookup failed: %w", err)
	}

	if !found {
		return nil, fmt.Errorf("TXT record not found, ensure you have added the record and that DNS has propagated")
	}

	now := time.Now()
	if err := s.repo.UpdateVerificationStatus(ctx, la.ID, true, &now); err != nil {
		return nil, err
	}

	la.Verified = true
	la.VerifiedAt = &now
	return &la, nil
}

func (s *Service) Unlink(ctx context.Context, userID uuid.UUID, provider string) error {
	return s.repo.Delete(ctx, userID, provider)
}

func generateVerifyToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
