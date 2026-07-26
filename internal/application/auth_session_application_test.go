package application

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"backend/core-server/internal/config"
	"backend/core-server/internal/domain"
	authinfra "backend/core-server/internal/infras/auth"
	"backend/core-server/internal/infras/cache"
	"backend/core-server/internal/model/entity"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type authSessionTestRig struct {
	service    *AuthService
	users      *stubUserRepo
	emailCodes *stubEmailCodeStore
	sessions   *cache.AuthStore
	tokens     *authinfra.JWTManager
	redis      *miniredis.Miniredis
	user       *entity.User
}

func newAuthSessionTestRig(t *testing.T) *authSessionTestRig {
	t.Helper()
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	sessions, err := cache.NewAuthStore(
		&cache.CacheClient{Cache: redisClient},
		&config.Config{Auth: config.AuthConfig{RefreshExpire: "168h"}},
	)
	if err != nil {
		t.Fatalf("NewAuthStore() error = %v", err)
	}
	passwordHash, err := authinfra.HashPassword("abc12345")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	user := &entity.User{
		ID: 1, Name: "test_user", Email: "user@qq.com", Password: passwordHash,
		Role: entity.UserRoleUser, Status: entity.UserStatusActive, AuthVersion: 1,
		Phone: "", Avatar: "", Sex: defaultUserSex,
	}
	users := &stubUserRepo{findByIDUser: user, findByNameUser: user, findByEmailUser: user}
	emailCodes := &stubEmailCodeStore{verifyValid: true}
	tokens := newTestJWTManager(t)
	service, err := NewAuthService(nil, users, emailCodes, sessions, &stubMailSender{}, tokens)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	return &authSessionTestRig{
		service: service, users: users, emailCodes: emailCodes, sessions: sessions,
		tokens: tokens, redis: redisServer, user: user,
	}
}

func TestAuthServiceLoginAndEmailLoginIssueExpectedTokens(t *testing.T) {
	tests := []struct {
		name  string
		login func(*authSessionTestRig) (*LoginResult, error)
	}{
		{name: "username", login: func(rig *authSessionTestRig) (*LoginResult, error) {
			return rig.service.Login(context.Background(), "test_user", "abc12345")
		}},
		{name: "email", login: func(rig *authSessionTestRig) (*LoginResult, error) {
			return rig.service.EmailLogin(context.Background(), "User@QQ.com", "abc12345")
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rig := newAuthSessionTestRig(t)
			result, err := test.login(rig)
			if err != nil {
				t.Fatalf("login error = %v", err)
			}
			if result.User.ID != rig.user.ID || result.Tokens.AccessExpiresIn != 20*60 || result.Tokens.RefreshExpiresIn != 7*24*60*60 {
				t.Fatalf("login result = %+v", result)
			}
			accessClaims, err := rig.tokens.ParseAccessToken(result.Tokens.AccessToken)
			if err != nil {
				t.Fatalf("ParseAccessToken() error = %v", err)
			}
			refreshClaims, err := rig.tokens.ParseRefreshToken(result.Tokens.RefreshToken)
			if err != nil {
				t.Fatalf("ParseRefreshToken() error = %v", err)
			}
			if accessClaims.SessionID != refreshClaims.SessionID || accessClaims.UserID != rig.user.ID {
				t.Fatalf("access claims = %+v, refresh claims = %+v", accessClaims, refreshClaims)
			}
		})
	}
}

func TestAuthServiceLoginDoesNotDistinguishMissingUserAndWrongPassword(t *testing.T) {
	rig := newAuthSessionTestRig(t)
	_, wrongPasswordErr := rig.service.Login(context.Background(), "test_user", "wrong123")
	rig.users.findByNameUser = nil
	_, missingUserErr := rig.service.Login(context.Background(), "missing", "wrong123")
	if !errors.Is(wrongPasswordErr, ErrInvalidCredentials) || !errors.Is(missingUserErr, ErrInvalidCredentials) {
		t.Fatalf("wrong password error = %v, missing user error = %v", wrongPasswordErr, missingUserErr)
	}
}

func TestAuthServiceSecondLoginFails(t *testing.T) {
	rig := newAuthSessionTestRig(t)
	if _, err := rig.service.Login(context.Background(), "test_user", "abc12345"); err != nil {
		t.Fatalf("first Login() error = %v", err)
	}
	if _, err := rig.service.Login(context.Background(), "test_user", "abc12345"); !errors.Is(err, ErrActiveSession) {
		t.Fatalf("second Login() error = %v, want %v", err, ErrActiveSession)
	}
}

func TestAuthServiceConcurrentLoginOnlyOneSucceeds(t *testing.T) {
	rig := newAuthSessionTestRig(t)
	const attempts = 16
	var successes int64
	var activeSessions int64
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := rig.service.Login(context.Background(), "test_user", "abc12345")
			switch {
			case err == nil:
				atomic.AddInt64(&successes, 1)
			case errors.Is(err, ErrActiveSession):
				atomic.AddInt64(&activeSessions, 1)
			default:
				t.Errorf("Login() error = %v", err)
			}
		}()
	}
	wg.Wait()
	if successes != 1 || activeSessions != attempts-1 {
		t.Fatalf("successes = %d, active sessions = %d", successes, activeSessions)
	}
}

func TestAuthServiceRefreshRotatesTokenAndRenewsTTL(t *testing.T) {
	rig := newAuthSessionTestRig(t)
	login, err := rig.service.Login(context.Background(), "test_user", "abc12345")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	rig.redis.FastForward(time.Hour)
	refreshed, err := rig.service.RefreshToken(context.Background(), login.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if refreshed.RefreshToken == login.Tokens.RefreshToken || refreshed.AccessToken == login.Tokens.AccessToken {
		t.Fatal("RefreshToken() did not rotate both tokens")
	}
	if ttl := rig.redis.TTL("auth:session:1"); ttl != 7*24*time.Hour {
		t.Fatalf("refreshed session TTL = %v", ttl)
	}
	refreshClaims, err := rig.tokens.ParseRefreshToken(refreshed.RefreshToken)
	if err != nil {
		t.Fatalf("ParseRefreshToken(refreshed) error = %v", err)
	}
	if ttl := rig.redis.TTL("auth:refresh:" + refreshClaims.ID); ttl != 7*24*time.Hour {
		t.Fatalf("refreshed JTI TTL = %v", ttl)
	}
	if _, err := rig.service.RefreshToken(context.Background(), login.Tokens.RefreshToken); !errors.Is(err, ErrRefreshInvalid) {
		t.Fatalf("reused old RefreshToken() error = %v", err)
	}
	if _, err := rig.service.ValidateAccess(context.Background(), refreshed.AccessToken); err != nil {
		t.Fatalf("ValidateAccess(new access) error = %v", err)
	}
}

func TestAuthServiceConcurrentRefreshOnlyOneSucceeds(t *testing.T) {
	rig := newAuthSessionTestRig(t)
	login, err := rig.service.Login(context.Background(), "test_user", "abc12345")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	const attempts = 16
	var successes int64
	var invalid int64
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, refreshErr := rig.service.RefreshToken(context.Background(), login.Tokens.RefreshToken)
			switch {
			case refreshErr == nil:
				atomic.AddInt64(&successes, 1)
			case errors.Is(refreshErr, ErrRefreshInvalid):
				atomic.AddInt64(&invalid, 1)
			default:
				t.Errorf("RefreshToken() error = %v", refreshErr)
			}
		}()
	}
	wg.Wait()
	if successes != 1 || invalid != attempts-1 {
		t.Fatalf("successes = %d, invalid = %d", successes, invalid)
	}
}

func TestAuthServiceLogoutInvalidatesAccessAndIsIdempotent(t *testing.T) {
	rig := newAuthSessionTestRig(t)
	login, err := rig.service.Login(context.Background(), "test_user", "abc12345")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if _, err := rig.service.ValidateAccess(context.Background(), login.Tokens.AccessToken); err != nil {
		t.Fatalf("ValidateAccess() before logout error = %v", err)
	}
	if err := rig.service.Logout(context.Background(), login.Tokens.RefreshToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if _, err := rig.service.ValidateAccess(context.Background(), login.Tokens.AccessToken); !errors.Is(err, ErrAccessInvalid) {
		t.Fatalf("ValidateAccess() after logout error = %v", err)
	}
	if _, err := rig.service.RefreshToken(context.Background(), login.Tokens.RefreshToken); !errors.Is(err, ErrRefreshInvalid) {
		t.Fatalf("RefreshToken() after logout error = %v", err)
	}
	if err := rig.service.Logout(context.Background(), login.Tokens.RefreshToken); err != nil {
		t.Fatalf("repeated Logout() error = %v", err)
	}
}

func TestAuthServiceOldRefreshCannotLogoutRotatedSession(t *testing.T) {
	rig := newAuthSessionTestRig(t)
	login, err := rig.service.Login(context.Background(), "test_user", "abc12345")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	refreshed, err := rig.service.RefreshToken(context.Background(), login.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if err := rig.service.Logout(context.Background(), login.Tokens.RefreshToken); err != nil {
		t.Fatalf("Logout(old refresh) error = %v", err)
	}
	if _, err := rig.service.ValidateAccess(context.Background(), refreshed.AccessToken); err != nil {
		t.Fatalf("old refresh deleted rotated session: %v", err)
	}
}

type failingClearSessionStore struct {
	domain.AuthSessionStore
}

func (f failingClearSessionStore) ClearUserSession(context.Context, uint64) error {
	return errors.New("redis cleanup failed")
}

func TestAuthServiceResetAuthVersionInvalidatesOldTokensWhenRedisCleanupFails(t *testing.T) {
	rig := newAuthSessionTestRig(t)
	login, err := rig.service.Login(context.Background(), "test_user", "abc12345")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	resetService, err := NewAuthService(
		nil,
		rig.users,
		rig.emailCodes,
		failingClearSessionStore{AuthSessionStore: rig.sessions},
		&stubMailSender{},
		rig.tokens,
	)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	if err := resetService.ResetPasswordByEmail(context.Background(), ResetPasswordByEmailInput{
		Email: "user@qq.com", EmailCode: "123456", NewPassword: "newabc123", ConfirmPassword: "newabc123",
	}); err != nil {
		t.Fatalf("ResetPasswordByEmail() error = %v", err)
	}
	if rig.user.AuthVersion != 2 {
		t.Fatalf("auth version = %d, want 2", rig.user.AuthVersion)
	}
	if _, err := rig.service.ValidateAccess(context.Background(), login.Tokens.AccessToken); !errors.Is(err, ErrAccessInvalid) {
		t.Fatalf("old access after reset error = %v", err)
	}
	if _, err := rig.service.RefreshToken(context.Background(), login.Tokens.RefreshToken); !errors.Is(err, ErrRefreshInvalid) {
		t.Fatalf("old refresh after reset error = %v", err)
	}
}

func TestAuthServiceRejectsDisabledUser(t *testing.T) {
	rig := newAuthSessionTestRig(t)
	rig.user.Status = "disabled"
	if _, err := rig.service.Login(context.Background(), "test_user", "abc12345"); !errors.Is(err, ErrUserDisabled) {
		t.Fatalf("Login(disabled) error = %v", err)
	}
}

func TestAuthServiceRejectsTokensAfterUserIsDisabled(t *testing.T) {
	rig := newAuthSessionTestRig(t)
	login, err := rig.service.Login(context.Background(), "test_user", "abc12345")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	rig.user.Status = "disabled"

	if _, err := rig.service.ValidateAccess(context.Background(), login.Tokens.AccessToken); !errors.Is(err, ErrUserDisabled) {
		t.Fatalf("ValidateAccess(disabled) error = %v", err)
	}
	if _, err := rig.service.RefreshToken(context.Background(), login.Tokens.RefreshToken); !errors.Is(err, ErrUserDisabled) {
		t.Fatalf("RefreshToken(disabled) error = %v", err)
	}
}
