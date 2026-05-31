package passkey

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/trackmycareer/app/internal/mfa"
)

// challengeTTL is how long a WebAuthn challenge remains valid.
const challengeTTL = 5 * time.Minute

// webauthnUser adapts the application's user data to satisfy the webauthn.User interface.
type webauthnUser struct {
	id          uuid.UUID
	email       string
	displayName string
	credentials []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte {
	b, _ := u.id.MarshalBinary()
	return b
}

func (u *webauthnUser) WebAuthnName() string {
	return u.email
}

func (u *webauthnUser) WebAuthnDisplayName() string {
	return u.displayName
}

func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

// Service handles WebAuthn passkey registration and authentication.
type Service struct {
	repo *Repository
	wa   *webauthn.WebAuthn
}

// NewService creates a new passkey Service. rpID is the relying party domain
// (e.g. "trackmy.career"), rpName is the display name, and rpOrigin is the
// fully qualified frontend origin (e.g. "https://trackmy.career").
func NewService(repo *Repository, rpID, rpName, rpOrigin string) (*Service, error) {
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: rpName,
		RPID:          rpID,
		RPOrigins:     []string{rpOrigin},
	})
	if err != nil {
		return nil, fmt.Errorf("initialising WebAuthn: %w", err)
	}
	return &Service{repo: repo, wa: wa}, nil
}

// toWebAuthnCredentials converts stored passkeys to the webauthn.Credential type.
func toWebAuthnCredentials(passkeys []mfa.Passkey) []webauthn.Credential {
	creds := make([]webauthn.Credential, len(passkeys))
	for i, p := range passkeys {
		transports := make([]protocol.AuthenticatorTransport, len(p.Transport))
		for j, t := range p.Transport {
			transports[j] = protocol.AuthenticatorTransport(t)
		}
		creds[i] = webauthn.Credential{
			ID:        p.CredentialID,
			PublicKey: p.PublicKey,
			Transport: transports,
			Flags: webauthn.CredentialFlags{
				BackupEligible: true,
				BackupState:    true,
			},
			Authenticator: webauthn.Authenticator{
				AAGUID:    p.AAGUID,
				SignCount: p.SignCount,
			},
		}
		if p.AttestationType != nil {
			creds[i].AttestationType = *p.AttestationType
		}
	}
	return creds
}

// BeginRegistration starts the WebAuthn credential registration ceremony.
func (s *Service) BeginRegistration(ctx context.Context, userID uuid.UUID, email, name string, existingCredentials []mfa.Passkey) (*protocol.CredentialCreation, error) {
	waCreds := toWebAuthnCredentials(existingCredentials)

	// Build exclude list so the authenticator does not re-register existing credentials.
	excludeList := make([]protocol.CredentialDescriptor, len(waCreds))
	for i, c := range waCreds {
		excludeList[i] = c.Descriptor()
	}

	user := &webauthnUser{
		id:          userID,
		email:       email,
		displayName: name,
		credentials: waCreds,
	}

	creation, session, err := s.wa.BeginRegistration(user, webauthn.WithExclusions(excludeList))
	if err != nil {
		return nil, fmt.Errorf("beginning WebAuthn registration: %w", err)
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("serialising session data: %w", err)
	}

	challenge := &mfa.WebAuthnChallenge{
		ID:          uuid.New(),
		UserID:      userID,
		Challenge:   []byte(session.Challenge),
		SessionData: sessionJSON,
		Operation:   "register",
		ExpiresAt:   time.Now().Add(challengeTTL),
	}

	if err := s.repo.StoreChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	return creation, nil
}

// CompleteRegistration finishes the WebAuthn credential registration ceremony.
func (s *Service) CompleteRegistration(ctx context.Context, userID uuid.UUID, email, name, passkeyName string, existingCredentials []mfa.Passkey, credentialJSON []byte) (*mfa.Passkey, error) {
	stored, err := s.repo.GetChallenge(ctx, userID, "register")
	if err != nil {
		return nil, fmt.Errorf("retrieving registration challenge: %w", err)
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(stored.SessionData, &session); err != nil {
		return nil, fmt.Errorf("deserialising session data: %w", err)
	}

	user := &webauthnUser{
		id:          userID,
		email:       email,
		displayName: name,
		credentials: toWebAuthnCredentials(existingCredentials),
	}

	parsed, err := protocol.ParseCredentialCreationResponseBytes(credentialJSON)
	if err != nil {
		return nil, fmt.Errorf("parsing WebAuthn attestation response: %w", err)
	}

	credential, err := s.wa.CreateCredential(user, session, parsed)
	if err != nil {
		return nil, fmt.Errorf("finishing WebAuthn registration: %w", err)
	}

	// Clean up the used challenge.
	_ = s.repo.DeleteChallenge(ctx, stored.ID)

	transports := make([]string, len(credential.Transport))
	for i, t := range credential.Transport {
		transports[i] = string(t)
	}

	p := &mfa.Passkey{
		ID:              uuid.New(),
		UserID:          userID,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: &credential.AttestationType,
		Transport:       transports,
		SignCount:       credential.Authenticator.SignCount,
		Name:            passkeyName,
		AAGUID:          credential.Authenticator.AAGUID,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

// BeginAuthentication starts the WebAuthn login assertion ceremony.
func (s *Service) BeginAuthentication(ctx context.Context, userID uuid.UUID, credentials []mfa.Passkey) (*protocol.CredentialAssertion, error) {
	waCreds := toWebAuthnCredentials(credentials)

	user := &webauthnUser{
		id:          userID,
		credentials: waCreds,
	}

	assertion, session, err := s.wa.BeginLogin(user)
	if err != nil {
		return nil, fmt.Errorf("beginning WebAuthn login: %w", err)
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("serialising session data: %w", err)
	}

	challenge := &mfa.WebAuthnChallenge{
		ID:          uuid.New(),
		UserID:      userID,
		Challenge:   []byte(session.Challenge),
		SessionData: sessionJSON,
		Operation:   "authenticate",
		ExpiresAt:   time.Now().Add(challengeTTL),
	}

	if err := s.repo.StoreChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	return assertion, nil
}

// CompleteAuthentication finishes the WebAuthn login assertion ceremony.
func (s *Service) CompleteAuthentication(ctx context.Context, userID uuid.UUID, credentials []mfa.Passkey, assertionJSON []byte) error {
	stored, err := s.repo.GetChallenge(ctx, userID, "authenticate")
	if err != nil {
		return fmt.Errorf("retrieving authentication challenge: %w", err)
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(stored.SessionData, &session); err != nil {
		return fmt.Errorf("deserialising session data: %w", err)
	}

	waCreds := toWebAuthnCredentials(credentials)

	user := &webauthnUser{
		id:          userID,
		credentials: waCreds,
	}

	parsed, err := protocol.ParseCredentialRequestResponseBytes(assertionJSON)
	if err != nil {
		return fmt.Errorf("parsing WebAuthn assertion response: %w", err)
	}

	credential, err := s.wa.ValidateLogin(user, session, parsed)
	if err != nil {
		return fmt.Errorf("finishing WebAuthn login: %w", err)
	}

	// Clean up the used challenge.
	_ = s.repo.DeleteChallenge(ctx, stored.ID)

	// Update sign count on the matched credential.
	for _, p := range credentials {
		if string(p.CredentialID) == string(credential.ID) {
			if updateErr := s.repo.UpdateSignCount(ctx, p.ID, credential.Authenticator.SignCount); updateErr != nil {
				return fmt.Errorf("updating sign count: %w", updateErr)
			}
			break
		}
	}

	return nil
}

// IsConfigured returns true if the user has at least one registered passkey.
func (s *Service) IsConfigured(ctx context.Context, userID uuid.UUID) (bool, error) {
	count, err := s.repo.Count(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

// ListByUserID returns all passkeys for the given user.
func (s *Service) ListByUserID(ctx context.Context, userID uuid.UUID) ([]mfa.Passkey, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// Rename updates the display name of a passkey.
func (s *Service) Rename(ctx context.Context, id, userID uuid.UUID, name string) error {
	return s.repo.Rename(ctx, id, userID, name)
}

// Delete removes a passkey.
func (s *Service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}

// Count returns the number of passkeys for a given user.
func (s *Service) Count(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.Count(ctx, userID)
}
