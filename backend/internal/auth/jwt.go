package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenDuration     = 15 * time.Minute
	RefreshTokenDuration    = 7 * 24 * time.Hour
	MFAPendingTokenDuration = 5 * time.Minute
)

type Claims struct {
	UserID        uuid.UUID `json:"user_id"`
	Email         string    `json:"email"`
	IsAdmin       bool      `json:"is_admin"`
	EmailVerified bool      `json:"email_verified"`
	TokenVersion  int       `json:"token_version"`
	MFAEnabled    bool      `json:"mfa_enabled"`
	TokenType     string    `json:"type,omitempty"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type JWTManager struct {
	secret []byte
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{secret: []byte(secret)}
}

func (m *JWTManager) GenerateTokenPair(userID uuid.UUID, email string, isAdmin, emailVerified bool, tokenVersion int, mfaEnabled bool) (TokenPair, error) {
	accessToken, err := m.generateToken(userID, email, isAdmin, emailVerified, tokenVersion, mfaEnabled, "", AccessTokenDuration)
	if err != nil {
		return TokenPair{}, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := m.generateToken(userID, email, isAdmin, emailVerified, tokenVersion, mfaEnabled, "", RefreshTokenDuration)
	if err != nil {
		return TokenPair{}, fmt.Errorf("generating refresh token: %w", err)
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GenerateMFAPendingToken creates a short-lived JWT (5 minutes) that indicates
// the user has passed credential verification but still needs to complete MFA.
// This token cannot be used as an access token; the auth middleware rejects it.
func (m *JWTManager) GenerateMFAPendingToken(userID uuid.UUID, email string) (string, error) {
	return m.generateToken(userID, email, false, false, 0, false, "mfa_pending", MFAPendingTokenDuration)
}

func (m *JWTManager) generateToken(userID uuid.UUID, email string, isAdmin, emailVerified bool, tokenVersion int, mfaEnabled bool, tokenType string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:        userID,
		Email:         email,
		IsAdmin:       isAdmin,
		EmailVerified: emailVerified,
		TokenVersion:  tokenVersion,
		MFAEnabled:    mfaEnabled,
		TokenType:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "trackmy-career",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
