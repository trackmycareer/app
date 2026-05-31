package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

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
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	Name     string `json:"name" binding:"required,max=255"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required"`
}

type RegisterResult struct {
	Tokens TokenPair
	UserID uuid.UUID
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (RegisterResult, error) {
	hash, err := password.Hash(req.Password)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("hashing password: %w", err)
	}

	u := &user.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Provider:     "email",
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return RegisterResult{}, fmt.Errorf("creating user: %w", err)
	}

	tokens, err := s.jwtManager.GenerateTokenPair(u.ID, u.Email, u.IsAdmin, u.EmailVerified, 0)
	if err != nil {
		return RegisterResult{}, err
	}

	return RegisterResult{Tokens: tokens, UserID: u.ID}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (TokenPair, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return TokenPair{}, fmt.Errorf("invalid email or password")
	}

	match, err := password.Verify(req.Password, u.PasswordHash)
	if err != nil {
		return TokenPair{}, fmt.Errorf("invalid email or password")
	}
	if !match {
		return TokenPair{}, fmt.Errorf("invalid email or password")
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

	return s.jwtManager.GenerateTokenPair(u.ID, u.Email, u.IsAdmin, u.EmailVerified, u.TokenVersion)
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

	return s.jwtManager.GenerateTokenPair(u.ID, u.Email, u.IsAdmin, u.EmailVerified, u.TokenVersion)
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.IncrementTokenVersion(ctx, userID)
}
