package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"backend/core-server/internal/domain"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestEmailCodeStore(t *testing.T) (*EmailCodeStore, *miniredis.Miniredis, *time.Time) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store, err := NewEmailCodeStore(&CacheClient{Cache: client})
	if err != nil {
		t.Fatalf("NewEmailCodeStore() error = %v", err)
	}
	now := time.Date(2030, 1, 1, 10, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }
	return store, server, &now
}

func advanceEmailStore(server *miniredis.Miniredis, now *time.Time, duration time.Duration) {
	server.FastForward(duration)
	*now = now.Add(duration)
}

func TestGenerateEmailCode(t *testing.T) {
	for i := 0; i < 100; i++ {
		code, err := GenerateEmailCode()
		if err != nil {
			t.Fatalf("GenerateEmailCode() error = %v", err)
		}
		if !isSixDigitCode(code) {
			t.Fatalf("GenerateEmailCode() = %q", code)
		}
	}
}

func TestEmailCodeTTLsAndExpiry(t *testing.T) {
	store, server, now := newTestEmailCodeStore(t)
	ctx := context.Background()
	code, err := store.CreateCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister)
	if err != nil {
		t.Fatalf("CreateCode() error = %v", err)
	}
	if ttl := server.TTL(emailCodeKey("user@qq.com", domain.EmailCodePurposeRegister)); ttl != 5*time.Minute {
		t.Fatalf("code TTL = %v", ttl)
	}
	if ttl := server.TTL(emailCooldownKey("user@qq.com")); ttl != 60*time.Second {
		t.Fatalf("cooldown TTL = %v", ttl)
	}
	if ttl := server.TTL(emailHourlyKey("user@qq.com", *now)); ttl != time.Hour {
		t.Fatalf("hourly counter TTL = %v", ttl)
	}

	advanceEmailStore(server, now, 5*time.Minute)
	valid, err := store.VerifyCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, code)
	if err != nil || valid {
		t.Fatalf("expired code validity = %v, %v", valid, err)
	}
}

func TestEmailCodePurposeIsolationAndSingleUse(t *testing.T) {
	store, server, now := newTestEmailCodeStore(t)
	ctx := context.Background()
	codes := []string{"111111", "222222"}
	store.generateCode = func() (string, error) {
		code := codes[0]
		codes = codes[1:]
		return code, nil
	}

	registerCode, err := store.CreateCode(ctx, "User@QQ.com", domain.EmailCodePurposeRegister)
	if err != nil {
		t.Fatalf("register CreateCode() error = %v", err)
	}
	advanceEmailStore(server, now, 61*time.Second)
	resetCode, err := store.CreateCode(ctx, "user@qq.com", domain.EmailCodePurposeResetPassword)
	if err != nil {
		t.Fatalf("reset CreateCode() error = %v", err)
	}

	valid, err := store.VerifyCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, registerCode)
	if err != nil || !valid {
		t.Fatalf("register VerifyCode() = %v, %v", valid, err)
	}
	valid, err = store.VerifyCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, registerCode)
	if err != nil || valid {
		t.Fatalf("reused register code = %v, %v", valid, err)
	}
	valid, err = store.VerifyCode(ctx, "user@qq.com", domain.EmailCodePurposeResetPassword, resetCode)
	if err != nil || !valid {
		t.Fatalf("reset VerifyCode() = %v, %v", valid, err)
	}
}

func TestEmailCodeNewCodeOverwritesOldCode(t *testing.T) {
	store, server, now := newTestEmailCodeStore(t)
	ctx := context.Background()
	codes := []string{"111111", "222222"}
	store.generateCode = func() (string, error) {
		code := codes[0]
		codes = codes[1:]
		return code, nil
	}
	oldCode, err := store.CreateCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister)
	if err != nil {
		t.Fatalf("first CreateCode() error = %v", err)
	}
	advanceEmailStore(server, now, 61*time.Second)
	newCode, err := store.CreateCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister)
	if err != nil {
		t.Fatalf("second CreateCode() error = %v", err)
	}
	valid, err := store.VerifyCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, oldCode)
	if err != nil || valid {
		t.Fatalf("old code validity = %v, %v", valid, err)
	}
	valid, err = store.VerifyCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, newCode)
	if err != nil || !valid {
		t.Fatalf("new code validity = %v, %v", valid, err)
	}
}

func TestEmailCodeCooldownAndHourlyLimit(t *testing.T) {
	store, server, now := newTestEmailCodeStore(t)
	ctx := context.Background()
	for i := 0; i < emailHourlyLimit; i++ {
		if _, err := store.CreateCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister); err != nil {
			t.Fatalf("CreateCode() attempt %d error = %v", i+1, err)
		}
		if _, err := store.CreateCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister); !errors.Is(err, ErrEmailCodeCooldown) {
			t.Fatalf("cooldown error = %v", err)
		}
		advanceEmailStore(server, now, 61*time.Second)
	}
	if _, err := store.CreateCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister); !errors.Is(err, ErrEmailCodeHourlyLimit) {
		t.Fatalf("hourly limit error = %v", err)
	}
}

func TestEmailCodeConcurrentCreateOnlyOneSucceeds(t *testing.T) {
	store, _, _ := newTestEmailCodeStore(t)
	ctx := context.Background()
	const attempts = 64
	var successes int64
	var cooldowns int64
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.CreateCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister)
			switch {
			case err == nil:
				atomic.AddInt64(&successes, 1)
			case errors.Is(err, ErrEmailCodeCooldown):
				atomic.AddInt64(&cooldowns, 1)
			default:
				t.Errorf("CreateCode() error = %v", err)
			}
		}()
	}
	wg.Wait()
	if successes != 1 || cooldowns != attempts-1 {
		t.Fatalf("successes = %d, cooldowns = %d", successes, cooldowns)
	}
}

func TestEmailCodeConcurrentVerifyOnlyOneSucceeds(t *testing.T) {
	store, _, _ := newTestEmailCodeStore(t)
	ctx := context.Background()
	code, err := store.CreateCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister)
	if err != nil {
		t.Fatalf("CreateCode() error = %v", err)
	}
	const attempts = 64
	var successes int64
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			valid, verifyErr := store.VerifyCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, code)
			if verifyErr != nil {
				t.Errorf("VerifyCode() error = %v", verifyErr)
				return
			}
			if valid {
				atomic.AddInt64(&successes, 1)
			}
		}()
	}
	wg.Wait()
	if successes != 1 {
		t.Fatalf("successful verifications = %d, want 1", successes)
	}
}

func TestEmailCodeCancelReservationRestoresRateLimitState(t *testing.T) {
	store, server, now := newTestEmailCodeStore(t)
	ctx := context.Background()
	reservation, err := store.ReserveCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister)
	if err != nil {
		t.Fatalf("ReserveCode() error = %v", err)
	}
	canceled, err := store.CancelCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, reservation)
	if err != nil || !canceled {
		t.Fatalf("CancelCode() = %v, %v", canceled, err)
	}
	if server.Exists(emailCodeKey("user@qq.com", domain.EmailCodePurposeRegister)) ||
		server.Exists(emailCooldownKey("user@qq.com")) ||
		server.Exists(emailHourlyKey("user@qq.com", *now)) {
		t.Fatal("canceled reservation left Redis rate-limit state behind")
	}
	if _, err := store.ReserveCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister); err != nil {
		t.Fatalf("immediate retry after cancellation error = %v", err)
	}
}

func TestEmailCodeStaleCancellationDoesNotDeleteNewReservation(t *testing.T) {
	store, server, now := newTestEmailCodeStore(t)
	ctx := context.Background()
	codes := []string{"111111", "222222"}
	store.generateCode = func() (string, error) {
		code := codes[0]
		codes = codes[1:]
		return code, nil
	}
	oldReservation, err := store.ReserveCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister)
	if err != nil {
		t.Fatalf("old ReserveCode() error = %v", err)
	}
	advanceEmailStore(server, now, 61*time.Second)
	newReservation, err := store.ReserveCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister)
	if err != nil {
		t.Fatalf("new ReserveCode() error = %v", err)
	}
	canceled, err := store.CancelCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, oldReservation)
	if err != nil || canceled {
		t.Fatalf("stale CancelCode() = %v, %v", canceled, err)
	}
	valid, err := store.VerifyCode(
		ctx,
		"user@qq.com",
		domain.EmailCodePurposeRegister,
		newReservation.Code,
	)
	if err != nil || !valid {
		t.Fatalf("new reservation validity = %v, %v", valid, err)
	}
}

func TestEmailCodeCancellationRequiresCurrentReservationID(t *testing.T) {
	store, _, _ := newTestEmailCodeStore(t)
	ctx := context.Background()
	reservation, err := store.ReserveCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister)
	if err != nil {
		t.Fatalf("ReserveCode() error = %v", err)
	}
	forged := *reservation
	forged.ID = "different-reservation"
	canceled, err := store.CancelCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, &forged)
	if err != nil || canceled {
		t.Fatalf("forged CancelCode() = %v, %v", canceled, err)
	}
	valid, err := store.VerifyCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, reservation.Code)
	if err != nil || !valid {
		t.Fatalf("current reservation validity = %v, %v", valid, err)
	}
}

func TestEmailCodeConcurrentCancellationAppliesOnceAndNeverGoesNegative(t *testing.T) {
	store, server, now := newTestEmailCodeStore(t)
	ctx := context.Background()
	reservation, err := store.ReserveCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister)
	if err != nil {
		t.Fatalf("ReserveCode() error = %v", err)
	}
	const attempts = 64
	var cancellations int64
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			canceled, cancelErr := store.CancelCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, reservation)
			if cancelErr != nil {
				t.Errorf("CancelCode() error = %v", cancelErr)
				return
			}
			if canceled {
				atomic.AddInt64(&cancellations, 1)
			}
		}()
	}
	wg.Wait()
	if cancellations != 1 {
		t.Fatalf("successful cancellations = %d, want 1", cancellations)
	}
	if server.Exists(emailHourlyKey("user@qq.com", *now)) {
		t.Fatal("hourly count still exists after one reservation was compensated")
	}
	canceled, err := store.CancelCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, reservation)
	if err != nil || canceled {
		t.Fatalf("repeated CancelCode() = %v, %v", canceled, err)
	}
}

func TestEmailCodeLateConcurrentFailuresDoNotBypassHourlyLimit(t *testing.T) {
	store, server, now := newTestEmailCodeStore(t)
	ctx := context.Background()
	reservations := make([]*domain.EmailCodeReservation, 0, emailHourlyLimit)
	for i := 0; i < emailHourlyLimit; i++ {
		reservation, err := store.ReserveCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister)
		if err != nil {
			t.Fatalf("ReserveCode() attempt %d error = %v", i+1, err)
		}
		reservations = append(reservations, reservation)
		if i < emailHourlyLimit-1 {
			advanceEmailStore(server, now, 61*time.Second)
		}
	}

	var wg sync.WaitGroup
	for _, reservation := range reservations[:len(reservations)-1] {
		reservation := reservation
		wg.Add(1)
		go func() {
			defer wg.Done()
			canceled, err := store.CancelCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister, reservation)
			if err != nil {
				t.Errorf("late CancelCode() error = %v", err)
			} else if canceled {
				t.Error("late cancellation unexpectedly changed the active reservation")
			}
		}()
	}
	wg.Wait()
	advanceEmailStore(server, now, 61*time.Second)
	if _, err := store.ReserveCode(ctx, "user@qq.com", domain.EmailCodePurposeRegister); !errors.Is(err, ErrEmailCodeHourlyLimit) {
		t.Fatalf("ReserveCode() after five delivered reservations error = %v", err)
	}
}
