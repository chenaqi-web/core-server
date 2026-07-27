package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"backend/core-server/internal/domain"

	"github.com/redis/go-redis/v9"
)

const (
	emailCodeTTL      = 5 * time.Minute
	emailCodeCooldown = 60 * time.Second
	emailHourlyLimit  = 5
)

var (
	ErrInvalidEmailCodePurpose = errors.New("invalid email code purpose")
	ErrEmailCodeCooldown       = errors.New("email code was sent too recently")
	ErrEmailCodeHourlyLimit    = errors.New("email code hourly limit reached")
)

type EmailCodeStore struct {
	client                *redis.Client
	now                   func() time.Time
	generateCode          func() (string, error)
	generateReservationID func() (string, error)
}

var createEmailCodeScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[2]) == 1 then
    return 1
end
local hourly_count = tonumber(redis.call('GET', KEYS[3]) or '0')
if hourly_count >= tonumber(ARGV[5]) then
    return 2
end
hourly_count = redis.call('INCR', KEYS[3])
if hourly_count == 1 then
	redis.call('EXPIRE', KEYS[3], ARGV[6])
end
redis.call('SET', KEYS[2], ARGV[2], 'EX', ARGV[4])
redis.call('HSET', KEYS[1], 'code', ARGV[1], 'reservation_id', ARGV[2])
redis.call('EXPIRE', KEYS[1], ARGV[3])
return 0
`)

var verifyEmailCodeScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'code') ~= ARGV[1] then
    return 0
end
redis.call('DEL', KEYS[1])
return 1
`)

var cancelEmailCodeScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'reservation_id') ~= ARGV[1] then
    return 0
end
if redis.call('HGET', KEYS[1], 'code') ~= ARGV[2] then
    return 0
end
redis.call('DEL', KEYS[1])
if redis.call('GET', KEYS[2]) == ARGV[1] then
    redis.call('DEL', KEYS[2])
end
local hourly_count = tonumber(redis.call('GET', KEYS[3]) or '0')
if hourly_count <= 1 then
    redis.call('DEL', KEYS[3])
else
    redis.call('DECR', KEYS[3])
end
return 1
`)

func NewEmailCodeStore(cacheClient *CacheClient) (*EmailCodeStore, error) {
	if cacheClient == nil || cacheClient.Cache == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	return &EmailCodeStore{
		client:                cacheClient.Cache,
		now:                   time.Now,
		generateCode:          GenerateEmailCode,
		generateReservationID: generateEmailCodeReservationID,
	}, nil
}

func GenerateEmailCode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generate email code: %w", err)
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func (s *EmailCodeStore) CreateCode(ctx context.Context, email string, purpose domain.EmailCodePurpose) (string, error) {
	reservation, err := s.ReserveCode(ctx, email, purpose)
	if err != nil {
		return "", err
	}
	return reservation.Code, nil
}

func (s *EmailCodeStore) ReserveCode(
	ctx context.Context,
	email string,
	purpose domain.EmailCodePurpose,
) (*domain.EmailCodeReservation, error) {
	if !purpose.Valid() {
		return nil, ErrInvalidEmailCodePurpose
	}
	email = normalizeEmail(email)
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	code, err := s.generateCode()
	if err != nil {
		return nil, err
	}
	reservationID, err := s.generateReservationID()
	if err != nil {
		return nil, err
	}

	now := s.now()
	hourBucket := now.Format("2006010215")
	hourlyTTL := now.Truncate(time.Hour).Add(time.Hour).Sub(now)
	if hourlyTTL < time.Second {
		hourlyTTL = time.Second
	}
	result, err := createEmailCodeScript.Run(
		ctx,
		s.client,
		[]string{
			emailCodeKey(email, purpose),
			emailCooldownKey(email),
			emailHourlyKey(email, now),
		},
		code,
		reservationID,
		int64(emailCodeTTL/time.Second),
		int64(emailCodeCooldown/time.Second),
		emailHourlyLimit,
		int64(hourlyTTL/time.Second),
	).Int64()
	if err != nil {
		return nil, fmt.Errorf("create email code: %w", err)
	}
	switch result {
	case 0:
		return &domain.EmailCodeReservation{
			Code:       code,
			ID:         reservationID,
			HourBucket: hourBucket,
		}, nil
	case 1:
		return nil, ErrEmailCodeCooldown
	case 2:
		return nil, ErrEmailCodeHourlyLimit
	default:
		return nil, fmt.Errorf("unexpected email code result: %d", result)
	}
}

func (s *EmailCodeStore) CancelCode(
	ctx context.Context,
	email string,
	purpose domain.EmailCodePurpose,
	reservation *domain.EmailCodeReservation,
) (bool, error) {
	if !purpose.Valid() {
		return false, ErrInvalidEmailCodePurpose
	}
	email = normalizeEmail(email)
	if email == "" || reservation == nil || reservation.ID == "" ||
		!isSixDigitCode(reservation.Code) || !isHourBucket(reservation.HourBucket) {
		return false, fmt.Errorf("invalid email code reservation")
	}
	result, err := cancelEmailCodeScript.Run(
		ctx,
		s.client,
		[]string{
			emailCodeKey(email, purpose),
			emailCooldownKey(email),
			emailHourlyKeyForBucket(email, reservation.HourBucket),
		},
		reservation.ID,
		reservation.Code,
	).Int64()
	if err != nil {
		return false, fmt.Errorf("cancel email code: %w", err)
	}
	return result == 1, nil
}

func (s *EmailCodeStore) VerifyCode(ctx context.Context, email string, purpose domain.EmailCodePurpose, code string) (bool, error) {
	if !purpose.Valid() {
		return false, ErrInvalidEmailCodePurpose
	}
	email = normalizeEmail(email)
	if email == "" || !isSixDigitCode(code) {
		return false, nil
	}
	result, err := verifyEmailCodeScript.Run(
		ctx,
		s.client,
		[]string{emailCodeKey(email, purpose)},
		code,
	).Int64()
	if err != nil {
		return false, fmt.Errorf("verify email code: %w", err)
	}
	return result == 1, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isSixDigitCode(code string) bool {
	if len(code) != 6 {
		return false
	}
	for _, character := range code {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func generateEmailCodeReservationID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate email code reservation ID: %w", err)
	}
	return hex.EncodeToString(value), nil
}

func isHourBucket(bucket string) bool {
	if len(bucket) != len("2006010215") {
		return false
	}
	_, err := time.Parse("2006010215", bucket)
	return err == nil
}

func emailCodeKey(email string, purpose domain.EmailCodePurpose) string {
	return "auth:email:code:" + string(purpose) + ":" + email
}

func emailCooldownKey(email string) string {
	return "auth:email:cooldown:" + email
}

func emailHourlyKey(email string, now time.Time) string {
	return emailHourlyKeyForBucket(email, now.Format("2006010215"))
}

func emailHourlyKeyForBucket(email, bucket string) string {
	return "auth:email:hourly:" + email + ":" + bucket
}
