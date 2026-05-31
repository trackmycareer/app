package mfa

import (
	"context"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
)

// TOTPService defines the interface for TOTP operations.
type TOTPService interface {
	Setup(ctx context.Context, userID uuid.UUID, email string) (uri string, secret string, err error)
	Verify(ctx context.Context, userID uuid.UUID, code string) error
	ValidateCode(ctx context.Context, userID uuid.UUID, code string) (bool, error)
	IsConfigured(ctx context.Context, userID uuid.UUID) (bool, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}

// PasskeyService defines the interface for passkey operations.
type PasskeyService interface {
	BeginRegistration(ctx context.Context, userID uuid.UUID, email, name string, existingCredentials []Passkey) (*protocol.CredentialCreation, error)
	CompleteRegistration(ctx context.Context, userID uuid.UUID, email, name, passkeyName string, existingCredentials []Passkey, credentialJSON []byte) (*Passkey, error)
	BeginAuthentication(ctx context.Context, userID uuid.UUID, credentials []Passkey) (*protocol.CredentialAssertion, error)
	CompleteAuthentication(ctx context.Context, userID uuid.UUID, credentials []Passkey, assertionJSON []byte) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]Passkey, error)
	Rename(ctx context.Context, id, userID uuid.UUID, name string) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	Count(ctx context.Context, userID uuid.UUID) (int, error)
	IsConfigured(ctx context.Context, userID uuid.UUID) (bool, error)
}

// BackupService defines the interface for backup code operations.
type BackupService interface {
	Generate(ctx context.Context, userID uuid.UUID) ([]string, error)
	VerifyAndConsume(ctx context.Context, userID uuid.UUID, code string) (bool, error)
	CountRemaining(ctx context.Context, userID uuid.UUID) (int, error)
	Regenerate(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// UserRepository defines the user operations needed by the MFA service.
type UserRepository interface {
	SetMFAEnabled(ctx context.Context, userID uuid.UUID, enabled bool) error
}

// Service orchestrates MFA operations across TOTP, passkey, and backup code sub-services.
type Service struct {
	totpSvc    TOTPService
	passkeySvc PasskeyService
	backupSvc  BackupService
	userRepo   UserRepository
}

// NewService creates a new MFA orchestration Service.
func NewService(totpSvc TOTPService, passkeySvc PasskeyService, backupSvc BackupService, userRepo UserRepository) *Service {
	return &Service{
		totpSvc:    totpSvc,
		passkeySvc: passkeySvc,
		backupSvc:  backupSvc,
		userRepo:   userRepo,
	}
}

// TOTP returns the TOTP sub-service.
func (s *Service) TOTP() TOTPService { return s.totpSvc }

// Passkey returns the passkey sub-service.
func (s *Service) Passkey() PasskeyService { return s.passkeySvc }

// Backup returns the backup code sub-service.
func (s *Service) Backup() BackupService { return s.backupSvc }

// GetStatus returns the aggregate MFA status for a user.
func (s *Service) GetStatus(ctx context.Context, userID uuid.UUID) (Status, error) {
	totpConfigured, err := s.totpSvc.IsConfigured(ctx, userID)
	if err != nil {
		return Status{}, fmt.Errorf("checking TOTP status: %w", err)
	}

	passkeyCount, err := s.passkeySvc.Count(ctx, userID)
	if err != nil {
		return Status{}, fmt.Errorf("checking passkey count: %w", err)
	}

	backupRemaining, err := s.backupSvc.CountRemaining(ctx, userID)
	if err != nil {
		return Status{}, fmt.Errorf("checking backup codes: %w", err)
	}

	methods := make([]Method, 0)
	if totpConfigured {
		methods = append(methods, MethodTOTP)
	}
	if passkeyCount > 0 {
		methods = append(methods, MethodPasskey)
	}

	status := Status{
		Enabled:              len(methods) > 0,
		Methods:              methods,
		PasskeyCount:         passkeyCount,
		BackupCodesRemaining: backupRemaining,
	}

	return status, nil
}

// HasAnyMethod returns true if the user has at least one verified MFA method.
func (s *Service) HasAnyMethod(ctx context.Context, userID uuid.UUID) (bool, error) {
	totpConfigured, err := s.totpSvc.IsConfigured(ctx, userID)
	if err != nil {
		return false, err
	}
	if totpConfigured {
		return true, nil
	}

	passkeyConfigured, err := s.passkeySvc.IsConfigured(ctx, userID)
	if err != nil {
		return false, err
	}
	return passkeyConfigured, nil
}

// UpdateMFAFlag checks all MFA methods and sets the mfa_enabled flag on the user accordingly.
func (s *Service) UpdateMFAFlag(ctx context.Context, userID uuid.UUID) error {
	hasAny, err := s.HasAnyMethod(ctx, userID)
	if err != nil {
		return fmt.Errorf("checking MFA methods: %w", err)
	}
	return s.userRepo.SetMFAEnabled(ctx, userID, hasAny)
}

// VerifyAnyMethod dispatches verification to the correct sub-service based on the method.
func (s *Service) VerifyAnyMethod(ctx context.Context, userID uuid.UUID, method Method, code string) error {
	switch method {
	case MethodTOTP:
		valid, err := s.totpSvc.ValidateCode(ctx, userID, code)
		if err != nil {
			return fmt.Errorf("validating TOTP code: %w", err)
		}
		if !valid {
			return fmt.Errorf("invalid verification code")
		}
		return nil

	case MethodBackup:
		valid, err := s.backupSvc.VerifyAndConsume(ctx, userID, code)
		if err != nil {
			return fmt.Errorf("verifying backup code: %w", err)
		}
		if !valid {
			return fmt.Errorf("invalid verification code")
		}
		return nil

	case MethodPasskey:
		// Passkey verification is handled separately through the WebAuthn flow
		// and cannot be reduced to a simple code string.
		return fmt.Errorf("passkey verification must use the WebAuthn flow")

	default:
		return fmt.Errorf("unsupported MFA method: %s", method)
	}
}

// DisableAll removes all MFA data for a user. Returns an error if the user is an admin.
func (s *Service) DisableAll(ctx context.Context, userID uuid.UUID, isAdmin bool) error {
	if isAdmin {
		return fmt.Errorf("administrators must have MFA enabled")
	}

	if err := s.totpSvc.Delete(ctx, userID); err != nil {
		return fmt.Errorf("deleting TOTP: %w", err)
	}

	// Passkey and backup code deletion is handled at the handler level
	// (iterating over passkeys) or via database cascade.
	if err := s.userRepo.SetMFAEnabled(ctx, userID, false); err != nil {
		return fmt.Errorf("updating MFA flag: %w", err)
	}

	return nil
}
