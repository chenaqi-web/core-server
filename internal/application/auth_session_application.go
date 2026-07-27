package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"backend/core-server/internal/domain"
	authinfra "backend/core-server/internal/infras/auth"
	"backend/core-server/internal/model/entity"
)

type AuthTokens struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresIn  int64
	RefreshExpiresIn int64
}

type LoginResult struct {
	User   *AuthUser
	Tokens *AuthTokens
}

type AccessIdentity struct {
	UserID      uint64
	SessionID   string
	Role        string
	Status      string
	AuthVersion uint64
}

type issuedTokenPair struct {
	tokens     *AuthTokens
	refreshJTI string
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, ErrInvalidCredentials
	}
	user, err := s.users.FindByName(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("find login user by name: %w", err)
	}
	return s.loginUser(ctx, user, password)
}

func (s *AuthService) EmailLogin(ctx context.Context, email, password string) (*LoginResult, error) {
	email, err := validateAndNormalizeEmail(email)
	if err != nil || password == "" {
		return nil, ErrInvalidCredentials
	}
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("find login user by email: %w", err)
	}
	return s.loginUser(ctx, user, password)
}

func (s *AuthService) RefreshToken(ctx context.Context, rawRefreshToken string) (*AuthTokens, error) {
	claims, err := s.tokens.ParseRefreshToken(strings.TrimSpace(rawRefreshToken))
	if err != nil {
		return nil, ErrRefreshInvalid
	}
	session, err := s.sessions.GetSession(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("get refresh session: %w", err)
	}
	if !sessionMatchesRefreshClaims(session, claims) {
		return nil, ErrRefreshInvalid
	}
	validRefresh, err := s.sessions.ValidateRefreshJTI(ctx, claims.UserID, claims.SessionID, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("validate refresh state: %w", err)
	}
	if !validRefresh {
		return nil, ErrRefreshInvalid
	}

	user, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("find refresh user: %w", err)
	}
	if user == nil || user.AuthVersion != claims.AuthVersion || user.AuthVersion != session.AuthVersion {
		return nil, ErrRefreshInvalid
	}
	if user.Status != entity.UserStatusActive {
		return nil, ErrUserDisabled
	}

	issued, err := s.issueTokenPair(user, claims.SessionID)
	if err != nil {
		return nil, err
	}
	rotated, err := s.sessions.RotateRefreshJTI(
		ctx,
		claims.UserID,
		claims.SessionID,
		claims.ID,
		issued.refreshJTI,
	)
	if err != nil {
		return nil, fmt.Errorf("rotate refresh state: %w", err)
	}
	if !rotated {
		return nil, ErrRefreshInvalid
	}
	return issued.tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	claims, err := s.tokens.ParseRefreshToken(strings.TrimSpace(rawRefreshToken))
	if err != nil {
		return ErrRefreshInvalid
	}
	if _, err := s.sessions.DeleteSessionIfMatch(ctx, claims.UserID, claims.SessionID, claims.ID); err != nil {
		return fmt.Errorf("delete logout session: %w", err)
	}
	return nil
}

func (s *AuthService) ValidateAccess(ctx context.Context, rawAccessToken string) (*AccessIdentity, error) {
	claims, err := s.tokens.ParseAccessToken(strings.TrimSpace(rawAccessToken))
	if err != nil {
		return nil, ErrAccessInvalid
	}
	session, err := s.sessions.GetSession(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("get access session: %w", err)
	}
	if !sessionMatchesAccessClaims(session, claims) {
		return nil, ErrAccessInvalid
	}

	user, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("find access user: %w", err)
	}
	if user == nil || user.AuthVersion != claims.AuthVersion || user.AuthVersion != session.AuthVersion {
		return nil, ErrAccessInvalid
	}
	if user.Status != entity.UserStatusActive {
		return nil, ErrUserDisabled
	}
	return &AccessIdentity{
		UserID:      user.ID,
		SessionID:   claims.SessionID,
		Role:        user.Role,
		Status:      user.Status,
		AuthVersion: user.AuthVersion,
	}, nil
}

func (s *AuthService) loginUser(ctx context.Context, user *entity.User, password string) (*LoginResult, error) {
	if user == nil || authinfra.VerifyPassword(user.Password, password) != nil {
		return nil, ErrInvalidCredentials
	}
	if user.Status != entity.UserStatusActive {
		return nil, ErrUserDisabled
	}
	if user.ID == 0 || user.AuthVersion == 0 {
		return nil, fmt.Errorf("login user has invalid authentication state")
	}

	sessionID, err := newSessionID()
	if err != nil {
		return nil, fmt.Errorf("generate login session ID: %w", err)
	}
	issued, err := s.issueTokenPair(user, sessionID)
	if err != nil {
		return nil, err
	}
	created, err := s.sessions.CreateSession(ctx, user.ID, domain.AuthSession{
		SessionID:   sessionID,
		RefreshJTI:  issued.refreshJTI,
		AuthVersion: user.AuthVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("create login session: %w", err)
	}
	if !created {
		return nil, ErrActiveSession
	}
	return &LoginResult{
		User:   authUserFromEntity(user),
		Tokens: issued.tokens,
	}, nil
}

func (s *AuthService) issueTokenPair(user *entity.User, sessionID string) (*issuedTokenPair, error) {
	refreshToken, refreshClaims, err := s.tokens.IssueRefreshToken(user.ID, sessionID, user.AuthVersion)
	if err != nil {
		return nil, fmt.Errorf("issue refresh token: %w", err)
	}
	accessToken, accessClaims, err := s.tokens.IssueAccessToken(user.ID, sessionID, user.AuthVersion)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}
	return &issuedTokenPair{
		tokens: &AuthTokens{
			AccessToken:      accessToken,
			RefreshToken:     refreshToken,
			AccessExpiresIn:  tokenLifetimeSeconds(accessClaims),
			RefreshExpiresIn: tokenLifetimeSeconds(refreshClaims),
		},
		refreshJTI: refreshClaims.ID,
	}, nil
}

func sessionMatchesRefreshClaims(session *domain.AuthSession, claims *authinfra.TokenClaims) bool {
	return session != nil && claims != nil &&
		session.SessionID == claims.SessionID &&
		session.RefreshJTI == claims.ID &&
		session.AuthVersion == claims.AuthVersion
}

func sessionMatchesAccessClaims(session *domain.AuthSession, claims *authinfra.TokenClaims) bool {
	return session != nil && claims != nil &&
		session.SessionID == claims.SessionID &&
		session.AuthVersion == claims.AuthVersion
}

func tokenLifetimeSeconds(claims *authinfra.TokenClaims) int64 {
	if claims == nil || claims.IssuedAt == nil || claims.ExpiresAt == nil {
		return 0
	}
	return int64(claims.ExpiresAt.Sub(claims.IssuedAt.Time) / time.Second)
}

func newSessionID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
