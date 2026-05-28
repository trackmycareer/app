package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/bhcloudlabs/trackmy-career/internal/user"
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
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (TokenPair, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return TokenPair{}, fmt.Errorf("hashing password: %w", err)
	}

	u := &user.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
		Provider:     "email",
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return TokenPair{}, fmt.Errorf("creating user: %w", err)
	}

	return s.jwtManager.GenerateTokenPair(u.ID, u.Email, u.IsAdmin)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (TokenPair, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return TokenPair{}, fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return TokenPair{}, fmt.Errorf("invalid email or password")
	}

	return s.jwtManager.GenerateTokenPair(u.ID, u.Email, u.IsAdmin)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return TokenPair{}, fmt.Errorf("invalid refresh token")
	}

	// Verify user still exists
	u, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("user not found")
	}

	return s.jwtManager.GenerateTokenPair(u.ID, u.Email, u.IsAdmin)
}
