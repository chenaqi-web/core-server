package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"backend/core-server/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
	minimumSecretLen = 32
)

var (
	ErrJWTConfig            = errors.New("invalid JWT configuration")
	ErrInvalidToken         = errors.New("invalid token")
	ErrExpiredToken         = errors.New("token expired")
	ErrInvalidSigningMethod = errors.New("invalid token signing method")
	ErrInvalidTokenType     = errors.New("invalid token type")
	ErrInvalidClaims        = errors.New("invalid token claims")
)

type TokenClaims struct {
	UserID      uint64 `json:"user_id"`
	SessionID   string `json:"session_id"`
	AuthVersion uint64 `json:"auth_version"`
	TokenType   string `json:"token_type"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	now           func() time.Time
}

func NewJWTManager(cfg *config.Config) (*JWTManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("%w: config is nil", ErrJWTConfig)
	}
	accessTTL, err := cfg.Auth.AccessDuration()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJWTConfig, err)
	}
	refreshTTL, err := cfg.Auth.RefreshDuration()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJWTConfig, err)
	}
	return NewJWTManagerWithConfig(
		cfg.Auth.JWTAccessSecret,
		cfg.Auth.JWTRefreshSecret,
		accessTTL,
		refreshTTL,
	)
}

func NewJWTManagerWithConfig(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) (*JWTManager, error) {
	if len([]byte(accessSecret)) < minimumSecretLen || len([]byte(refreshSecret)) < minimumSecretLen {
		return nil, fmt.Errorf("%w: each secret must contain at least %d bytes", ErrJWTConfig, minimumSecretLen)
	}
	if accessSecret == refreshSecret {
		return nil, fmt.Errorf("%w: access and refresh secrets must differ", ErrJWTConfig)
	}
	if accessTTL <= 0 || refreshTTL <= 0 {
		return nil, fmt.Errorf("%w: token durations must be positive", ErrJWTConfig)
	}
	return &JWTManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		now:           time.Now,
	}, nil
}

func (m *JWTManager) IssueAccessToken(userID uint64, sessionID string, authVersion uint64) (string, *TokenClaims, error) {
	return m.issueToken(userID, sessionID, authVersion, TokenTypeAccess, m.accessSecret, m.accessTTL)
}

func (m *JWTManager) IssueRefreshToken(userID uint64, sessionID string, authVersion uint64) (string, *TokenClaims, error) {
	return m.issueToken(userID, sessionID, authVersion, TokenTypeRefresh, m.refreshSecret, m.refreshTTL)
}

func (m *JWTManager) ParseAccessToken(token string) (*TokenClaims, error) {
	return m.parseToken(token, TokenTypeAccess, m.accessSecret)
}

func (m *JWTManager) ParseRefreshToken(token string) (*TokenClaims, error) {
	return m.parseToken(token, TokenTypeRefresh, m.refreshSecret)
}

func (m *JWTManager) issueToken(
	userID uint64,
	sessionID string,
	authVersion uint64,
	tokenType string,
	secret []byte,
	ttl time.Duration,
) (string, *TokenClaims, error) {
	now := m.now().UTC()
	jti, err := newTokenID()
	if err != nil {
		return "", nil, fmt.Errorf("generate token ID: %w", err)
	}
	claims := &TokenClaims{
		UserID:      userID,
		SessionID:   sessionID,
		AuthVersion: authVersion,
		TokenType:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	if err := validateClaims(claims, tokenType); err != nil {
		return "", nil, err
	}

	signedToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", nil, fmt.Errorf("sign token: %w", err)
	}
	return signedToken, claims, nil
}

func (m *JWTManager) parseToken(rawToken, expectedType string, secret []byte) (*TokenClaims, error) {
	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidSigningMethod
			}
			return secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithStrictDecoding(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrExpiredToken
		case errors.Is(err, ErrInvalidSigningMethod):
			return nil, ErrInvalidSigningMethod
		default:
			return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
		}
	}
	if token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if err := validateClaims(claims, expectedType); err != nil {
		return nil, err
	}
	return claims, nil
}

func validateClaims(claims *TokenClaims, expectedType string) error {
	if claims == nil || claims.UserID == 0 || claims.SessionID == "" || claims.AuthVersion == 0 {
		return ErrInvalidClaims
	}
	if claims.TokenType != expectedType {
		return ErrInvalidTokenType
	}
	if claims.ID == "" || claims.IssuedAt == nil || claims.ExpiresAt == nil {
		return ErrInvalidClaims
	}
	if !claims.ExpiresAt.After(claims.IssuedAt.Time) {
		return ErrInvalidClaims
	}
	return nil
}

func newTokenID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
