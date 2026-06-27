package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/legal"
	"github.com/trackmycareer/app/internal/password"
	"github.com/trackmycareer/app/internal/user"
)

type Service struct {
	userRepo   *user.Repository
	jwtManager *JWTManager
}

func NewService(userRepo *user.Repository, jwtManager *JWTManager) *Service {
	return &Service{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email,max=255"`
	Password    string `json:"password" binding:"required,min=8,max=128"`
	Name        string `json:"name" binding:"required,max=255"`
	AcceptTerms bool   `json:"accept_terms"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required"`
}

type RegisterResult struct {
	Tokens TokenPair
	UserID uuid.UUID
}

// LoginResult encapsulates the outcome of a login attempt. When the user has
// MFA enabled, TokenPair is nil and MFASession contains a short-lived JWT
// that must be presented to the MFA verification endpoint.
type LoginResult struct {
	TokenPair  *TokenPair // nil if MFA verification is required
	MFASession string     // set if MFA verification is required
	Methods    []string   // available MFA methods when MFA is required
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (RegisterResult, error) {
	hash, err := password.Hash(req.Password)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now()
	ver := legal.Version
	u := &user.User{
		ID:              uuid.New(),
		Email:           req.Email,
		PasswordHash:    hash,
		Name:            req.Name,
		Provider:        "email",
		TermsAcceptedAt: &now,
		TermsVersion:    &ver,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return RegisterResult{}, fmt.Errorf("creating user: %w", err)
	}

	tokens, err := s.jwtManager.GenerateTokenPair(u.ID, u.Email, u.IsAdmin, u.EmailVerified, 0, false)
	if err != nil {
		return RegisterResult{}, err
	}

	return RegisterResult{Tokens: tokens, UserID: u.ID}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResult, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return LoginResult{}, fmt.Errorf("invalid email or password")
	}

	match, err := password.Verify(req.Password, u.PasswordHash)
	if err != nil {
		return LoginResult{}, fmt.Errorf("invalid email or password")
	}
	if !match {
		return LoginResult{}, fmt.Errorf("invalid email or password")
	}

	// Transparently rehash bcrypt passwords to Argon2id on successful login.
	if password.NeedsRehash(u.PasswordHash) {
		newHash, err := password.Hash(req.Password)
		if err != nil {
			slog.Error("failed to rehash password to argon2id", "user_id", u.ID, "error", err)
		} else if err := s.userRepo.UpdatePassword(ctx, u.ID, newHash); err != nil {
			slog.Error("failed to persist rehashed password", "user_id", u.ID, "error", err)
		}
	}

	// If MFA is enabled, return a short-lived pending token instead of a full session.
	if u.MFAEnabled {
		mfaSession, err := s.jwtManager.GenerateMFAPendingToken(u.ID, u.Email)
		if err != nil {
			return LoginResult{}, fmt.Errorf("generating MFA session token: %w", err)
		}
		return LoginResult{
			MFASession: mfaSession,
			Methods:    []string{"totp", "passkey", "backup"},
		}, nil
	}

	tokens, err := s.jwtManager.GenerateTokenPair(u.ID, u.Email, u.IsAdmin, u.EmailVerified, u.TokenVersion, u.MFAEnabled)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{TokenPair: &tokens}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return TokenPair{}, fmt.Errorf("invalid refresh token")
	}

	u, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("user not found")
	}

	if claims.TokenVersion != u.TokenVersion {
		return TokenPair{}, fmt.Errorf("token has been revoked")
	}

	return s.jwtManager.GenerateTokenPair(u.ID, u.Email, u.IsAdmin, u.EmailVerified, u.TokenVersion, u.MFAEnabled)
}

// CompleteMFALogin validates an MFA-pending session token and generates a full
// token pair. The caller is responsible for verifying the MFA code before
// calling this method.
func (s *Service) CompleteMFALogin(ctx context.Context, mfaSessionToken string) (*TokenPair, uuid.UUID, error) {
	claims, err := s.jwtManager.ValidateToken(mfaSessionToken)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("MFA session expired, please log in again")
	}

	if claims.TokenType != "mfa_pending" {
		return nil, uuid.Nil, fmt.Errorf("invalid MFA session token")
	}

	u, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("user not found")
	}

	tokens, err := s.jwtManager.GenerateTokenPair(u.ID, u.Email, u.IsAdmin, u.EmailVerified, u.TokenVersion, u.MFAEnabled)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("generating token pair: %w", err)
	}

	return &tokens, u.ID, nil
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.IncrementTokenVersion(ctx, userID)
}
