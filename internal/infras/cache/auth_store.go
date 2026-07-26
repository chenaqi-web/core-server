package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"backend/core-server/internal/config"

	"github.com/redis/go-redis/v9"
)

const (
	authSessionKeyPrefix = "auth:session:"
	authRefreshKeyPrefix = "auth:refresh:"
)

var ErrInvalidAuthSession = errors.New("invalid auth session")

type AuthSession struct {
	SessionID   string
	RefreshJTI  string
	AuthVersion uint64
}

type AuthStore struct {
	client *redis.Client
	ttl    time.Duration
}

var createSessionScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 1 or redis.call('EXISTS', KEYS[2]) == 1 then
    return 0
end
redis.call('HSET', KEYS[1],
    'session_id', ARGV[1],
    'refresh_jti', ARGV[2],
    'auth_version', ARGV[3])
redis.call('EXPIRE', KEYS[1], ARGV[5])
redis.call('SET', KEYS[2], ARGV[4], 'EX', ARGV[5])
return 1
`)

var rotateRefreshScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then
    return 0
end
if redis.call('HGET', KEYS[1], 'session_id') ~= ARGV[1] then
    return 0
end
if redis.call('HGET', KEYS[1], 'refresh_jti') ~= ARGV[2] then
    return 0
end
if redis.call('GET', KEYS[2]) ~= ARGV[4] then
    return 0
end
if redis.call('EXISTS', KEYS[3]) == 1 then
    return 0
end
redis.call('DEL', KEYS[2])
redis.call('SET', KEYS[3], ARGV[4], 'EX', ARGV[5])
redis.call('HSET', KEYS[1], 'refresh_jti', ARGV[3])
redis.call('EXPIRE', KEYS[1], ARGV[5])
return 1
`)

var deleteSessionScript = redis.NewScript(`
local refresh_jti = redis.call('HGET', KEYS[1], 'refresh_jti')
redis.call('DEL', KEYS[1])
if refresh_jti then
    redis.call('DEL', ARGV[1] .. refresh_jti)
end
return 1
`)

func NewAuthStore(cacheClient *CacheClient, cfg *config.Config) (*AuthStore, error) {
	if cfg == nil {
		return nil, fmt.Errorf("auth store config is nil")
	}
	ttl, err := cfg.Auth.RefreshDuration()
	if err != nil {
		return nil, err
	}
	return newAuthStore(cacheClient, ttl)
}

func newAuthStore(cacheClient *CacheClient, ttl time.Duration) (*AuthStore, error) {
	if cacheClient == nil || cacheClient.Cache == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("auth store TTL must be positive")
	}
	return &AuthStore{client: cacheClient.Cache, ttl: ttl}, nil
}

func (s *AuthStore) CreateSession(ctx context.Context, userID uint64, session AuthSession) (bool, error) {
	if err := validateAuthSession(userID, session); err != nil {
		return false, err
	}
	result, err := createSessionScript.Run(
		ctx,
		s.client,
		[]string{sessionKey(userID), refreshKey(session.RefreshJTI)},
		session.SessionID,
		session.RefreshJTI,
		strconv.FormatUint(session.AuthVersion, 10),
		refreshIdentity(userID, session.SessionID),
		int64(s.ttl/time.Second),
	).Int64()
	if err != nil {
		return false, fmt.Errorf("create auth session: %w", err)
	}
	return result == 1, nil
}

func (s *AuthStore) GetSession(ctx context.Context, userID uint64) (*AuthSession, error) {
	if userID == 0 {
		return nil, ErrInvalidAuthSession
	}
	values, err := s.client.HGetAll(ctx, sessionKey(userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("get auth session: %w", err)
	}
	if len(values) == 0 {
		return nil, nil
	}
	authVersion, err := strconv.ParseUint(values["auth_version"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse auth session version: %w", err)
	}
	session := &AuthSession{
		SessionID:   values["session_id"],
		RefreshJTI:  values["refresh_jti"],
		AuthVersion: authVersion,
	}
	if err := validateAuthSession(userID, *session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *AuthStore) DeleteSession(ctx context.Context, userID uint64) error {
	if userID == 0 {
		return ErrInvalidAuthSession
	}
	if err := deleteSessionScript.Run(
		ctx,
		s.client,
		[]string{sessionKey(userID)},
		authRefreshKeyPrefix,
	).Err(); err != nil {
		return fmt.Errorf("delete auth session: %w", err)
	}
	return nil
}

func (s *AuthStore) ClearUserSession(ctx context.Context, userID uint64) error {
	return s.DeleteSession(ctx, userID)
}

func (s *AuthStore) SaveRefreshJTI(ctx context.Context, userID uint64, sessionID, refreshJTI string) (bool, error) {
	if userID == 0 || sessionID == "" || refreshJTI == "" {
		return false, ErrInvalidAuthSession
	}
	created, err := s.client.SetNX(
		ctx,
		refreshKey(refreshJTI),
		refreshIdentity(userID, sessionID),
		s.ttl,
	).Result()
	if err != nil {
		return false, fmt.Errorf("save refresh JTI: %w", err)
	}
	return created, nil
}

func (s *AuthStore) ValidateRefreshJTI(ctx context.Context, userID uint64, sessionID, refreshJTI string) (bool, error) {
	if userID == 0 || sessionID == "" || refreshJTI == "" {
		return false, ErrInvalidAuthSession
	}
	value, err := s.client.Get(ctx, refreshKey(refreshJTI)).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("validate refresh JTI: %w", err)
	}
	return value == refreshIdentity(userID, sessionID), nil
}

func (s *AuthStore) RotateRefreshJTI(
	ctx context.Context,
	userID uint64,
	sessionID, oldRefreshJTI, newRefreshJTI string,
) (bool, error) {
	if userID == 0 || sessionID == "" || oldRefreshJTI == "" || newRefreshJTI == "" || oldRefreshJTI == newRefreshJTI {
		return false, ErrInvalidAuthSession
	}
	result, err := rotateRefreshScript.Run(
		ctx,
		s.client,
		[]string{sessionKey(userID), refreshKey(oldRefreshJTI), refreshKey(newRefreshJTI)},
		sessionID,
		oldRefreshJTI,
		newRefreshJTI,
		refreshIdentity(userID, sessionID),
		int64(s.ttl/time.Second),
	).Int64()
	if err != nil {
		return false, fmt.Errorf("rotate refresh JTI: %w", err)
	}
	return result == 1, nil
}

func validateAuthSession(userID uint64, session AuthSession) error {
	if userID == 0 || session.SessionID == "" || session.RefreshJTI == "" || session.AuthVersion == 0 {
		return ErrInvalidAuthSession
	}
	return nil
}

func sessionKey(userID uint64) string {
	return authSessionKeyPrefix + strconv.FormatUint(userID, 10)
}

func refreshKey(refreshJTI string) string {
	return authRefreshKeyPrefix + refreshJTI
}

func refreshIdentity(userID uint64, sessionID string) string {
	return strconv.FormatUint(userID, 10) + ":" + sessionID
}
