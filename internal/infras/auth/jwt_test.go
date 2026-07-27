package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testAccessSecret  = "access-secret-0123456789-ABCDEFGHIJ"
	testRefreshSecret = "refresh-secret-0123456789-ABCDEFGH"
)

func newTestJWTManager(t *testing.T) *JWTManager {
	t.Helper()
	manager, err := NewJWTManagerWithConfig(testAccessSecret, testRefreshSecret, 20*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("NewJWTManagerWithConfig() error = %v", err)
	}
	manager.now = func() time.Time { return time.Unix(1_800_000_000, 0).UTC() }
	return manager
}

func TestJWTManagerIssueAndParseTokens(t *testing.T) {
	manager := newTestJWTManager(t)

	accessToken, accessClaims, err := manager.IssueAccessToken(42, "session-1", 3)
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}
	parsedAccess, err := manager.ParseAccessToken(accessToken)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}
	assertClaims(t, parsedAccess, 42, "session-1", 3, TokenTypeAccess)
	if accessClaims.ExpiresAt.Sub(accessClaims.IssuedAt.Time) != 20*time.Minute {
		t.Fatalf("access TTL = %v", accessClaims.ExpiresAt.Sub(accessClaims.IssuedAt.Time))
	}

	refreshToken, refreshClaims, err := manager.IssueRefreshToken(42, "session-1", 3)
	if err != nil {
		t.Fatalf("IssueRefreshToken() error = %v", err)
	}
	parsedRefresh, err := manager.ParseRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("ParseRefreshToken() error = %v", err)
	}
	assertClaims(t, parsedRefresh, 42, "session-1", 3, TokenTypeRefresh)
	if refreshClaims.ExpiresAt.Sub(refreshClaims.IssuedAt.Time) != 7*24*time.Hour {
		t.Fatalf("refresh TTL = %v", refreshClaims.ExpiresAt.Sub(refreshClaims.IssuedAt.Time))
	}
	if accessClaims.ID == refreshClaims.ID {
		t.Fatal("access and refresh tokens reused the same JTI")
	}
}

func TestJWTManagerRejectsWrongTokenType(t *testing.T) {
	manager := newTestJWTManager(t)
	refreshToken, _, err := manager.IssueRefreshToken(1, "session", 1)
	if err != nil {
		t.Fatalf("IssueRefreshToken() error = %v", err)
	}
	if _, err := manager.ParseAccessToken(refreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("ParseAccessToken() error = %v, want invalid token", err)
	}

	claims := validClaims(manager.now(), TokenTypeRefresh)
	raw := signClaims(t, claims, testAccessSecret, jwt.SigningMethodHS256)
	if _, err := manager.ParseAccessToken(raw); !errors.Is(err, ErrInvalidTokenType) {
		t.Fatalf("ParseAccessToken() error = %v, want %v", err, ErrInvalidTokenType)
	}
}

func TestJWTManagerRejectsWrongAlgorithm(t *testing.T) {
	manager := newTestJWTManager(t)
	raw := signClaims(t, validClaims(manager.now(), TokenTypeAccess), testAccessSecret, jwt.SigningMethodHS384)
	if _, err := manager.ParseAccessToken(raw); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("ParseAccessToken() error = %v, want invalid token", err)
	}
}

func TestJWTManagerRejectsWrongSecret(t *testing.T) {
	manager := newTestJWTManager(t)
	raw := signClaims(t, validClaims(manager.now(), TokenTypeAccess), "different-access-secret-0123456789", jwt.SigningMethodHS256)
	if _, err := manager.ParseAccessToken(raw); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("ParseAccessToken() error = %v, want invalid token", err)
	}
}

func TestJWTManagerRejectsExpiredToken(t *testing.T) {
	manager := newTestJWTManager(t)
	claims := validClaims(manager.now().Add(-time.Hour), TokenTypeAccess)
	claims.ExpiresAt = jwt.NewNumericDate(manager.now().Add(-time.Minute))
	raw := signClaims(t, claims, testAccessSecret, jwt.SigningMethodHS256)
	if _, err := manager.ParseAccessToken(raw); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("ParseAccessToken() error = %v, want %v", err, ErrExpiredToken)
	}
}

func TestJWTManagerRejectsMissingRequiredClaims(t *testing.T) {
	manager := newTestJWTManager(t)
	tests := []struct {
		name   string
		mutate func(*TokenClaims)
	}{
		{name: "user ID", mutate: func(c *TokenClaims) { c.UserID = 0 }},
		{name: "session ID", mutate: func(c *TokenClaims) { c.SessionID = "" }},
		{name: "auth version", mutate: func(c *TokenClaims) { c.AuthVersion = 0 }},
		{name: "JTI", mutate: func(c *TokenClaims) { c.ID = "" }},
		{name: "issued at", mutate: func(c *TokenClaims) { c.IssuedAt = nil }},
		{name: "expires at", mutate: func(c *TokenClaims) { c.ExpiresAt = nil }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := validClaims(manager.now(), TokenTypeAccess)
			test.mutate(claims)
			raw := signClaims(t, claims, testAccessSecret, jwt.SigningMethodHS256)
			if _, err := manager.ParseAccessToken(raw); err == nil {
				t.Fatal("ParseAccessToken() expected an error")
			}
		})
	}
}

func TestJWTManagerRejectsInvalidConfiguration(t *testing.T) {
	if _, err := NewJWTManagerWithConfig("short", testRefreshSecret, time.Minute, time.Hour); !errors.Is(err, ErrJWTConfig) {
		t.Fatalf("short secret error = %v", err)
	}
	if _, err := NewJWTManagerWithConfig(testAccessSecret, testAccessSecret, time.Minute, time.Hour); !errors.Is(err, ErrJWTConfig) {
		t.Fatalf("same secret error = %v", err)
	}
	if _, err := NewJWTManagerWithConfig(testAccessSecret, testRefreshSecret, 0, time.Hour); !errors.Is(err, ErrJWTConfig) {
		t.Fatalf("duration error = %v", err)
	}
}

func validClaims(now time.Time, tokenType string) *TokenClaims {
	return &TokenClaims{
		UserID:      1,
		SessionID:   "session",
		AuthVersion: 1,
		TokenType:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "jti",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}
}

func signClaims(t *testing.T, claims *TokenClaims, secret string, method jwt.SigningMethod) string {
	t.Helper()
	raw, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return raw
}

func assertClaims(t *testing.T, claims *TokenClaims, userID uint64, sessionID string, authVersion uint64, tokenType string) {
	t.Helper()
	if claims.UserID != userID || claims.SessionID != sessionID || claims.AuthVersion != authVersion || claims.TokenType != tokenType {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.ID == "" || claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatalf("missing registered claims: %+v", claims.RegisteredClaims)
	}
}
