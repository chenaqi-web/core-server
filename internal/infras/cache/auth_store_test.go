package cache

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestAuthStore(t *testing.T) (*AuthStore, *miniredis.Miniredis) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store, err := newAuthStore(&CacheClient{Cache: client}, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("newAuthStore() error = %v", err)
	}
	return store, server
}

func TestAuthStoreSessionLifecycle(t *testing.T) {
	store, server := newTestAuthStore(t)
	ctx := context.Background()
	session := AuthSession{SessionID: "session-1", RefreshJTI: "refresh-1", AuthVersion: 2}

	created, err := store.CreateSession(ctx, 42, session)
	if err != nil || !created {
		t.Fatalf("CreateSession() = %v, %v", created, err)
	}
	loaded, err := store.GetSession(ctx, 42)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if loaded == nil || *loaded != session {
		t.Fatalf("GetSession() = %+v, want %+v", loaded, session)
	}
	valid, err := store.ValidateRefreshJTI(ctx, 42, session.SessionID, session.RefreshJTI)
	if err != nil || !valid {
		t.Fatalf("ValidateRefreshJTI() = %v, %v", valid, err)
	}
	if ttl := server.TTL(sessionKey(42)); ttl != 7*24*time.Hour {
		t.Fatalf("session TTL = %v", ttl)
	}
	if ttl := server.TTL(refreshKey(session.RefreshJTI)); ttl != 7*24*time.Hour {
		t.Fatalf("refresh JTI TTL = %v", ttl)
	}

	if err := store.DeleteSession(ctx, 42); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	loaded, err = store.GetSession(ctx, 42)
	if err != nil || loaded != nil {
		t.Fatalf("GetSession() after delete = %+v, %v", loaded, err)
	}
	valid, err = store.ValidateRefreshJTI(ctx, 42, session.SessionID, session.RefreshJTI)
	if err != nil || valid {
		t.Fatalf("ValidateRefreshJTI() after delete = %v, %v", valid, err)
	}
}

func TestAuthStoreConcurrentLoginOnlyOneSucceeds(t *testing.T) {
	store, _ := newTestAuthStore(t)
	ctx := context.Background()
	const attempts = 64
	var successes int64
	var wg sync.WaitGroup

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			created, err := store.CreateSession(ctx, 7, AuthSession{
				SessionID:   fmt.Sprintf("session-%d", index),
				RefreshJTI:  fmt.Sprintf("refresh-%d", index),
				AuthVersion: 1,
			})
			if err != nil {
				t.Errorf("CreateSession() error = %v", err)
				return
			}
			if created {
				atomic.AddInt64(&successes, 1)
			}
		}(i)
	}
	wg.Wait()
	if successes != 1 {
		t.Fatalf("successful concurrent logins = %d, want 1", successes)
	}
}

func TestAuthStoreConcurrentRefreshOnlyOneSucceeds(t *testing.T) {
	store, _ := newTestAuthStore(t)
	ctx := context.Background()
	initial := AuthSession{SessionID: "session", RefreshJTI: "old-refresh", AuthVersion: 1}
	created, err := store.CreateSession(ctx, 9, initial)
	if err != nil || !created {
		t.Fatalf("CreateSession() = %v, %v", created, err)
	}

	const attempts = 64
	var successes int64
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			rotated, err := store.RotateRefreshJTI(ctx, 9, initial.SessionID, initial.RefreshJTI, fmt.Sprintf("new-%d", index))
			if err != nil {
				t.Errorf("RotateRefreshJTI() error = %v", err)
				return
			}
			if rotated {
				atomic.AddInt64(&successes, 1)
			}
		}(i)
	}
	wg.Wait()
	if successes != 1 {
		t.Fatalf("successful concurrent refreshes = %d, want 1", successes)
	}

	oldValid, err := store.ValidateRefreshJTI(ctx, 9, initial.SessionID, initial.RefreshJTI)
	if err != nil || oldValid {
		t.Fatalf("old refresh validity = %v, %v", oldValid, err)
	}
	loaded, err := store.GetSession(ctx, 9)
	if err != nil || loaded == nil || loaded.RefreshJTI == initial.RefreshJTI {
		t.Fatalf("rotated session = %+v, %v", loaded, err)
	}
	newValid, err := store.ValidateRefreshJTI(ctx, 9, initial.SessionID, loaded.RefreshJTI)
	if err != nil || !newValid {
		t.Fatalf("new refresh validity = %v, %v", newValid, err)
	}
}

func TestAuthStoreSaveRefreshJTIDoesNotOverwrite(t *testing.T) {
	store, _ := newTestAuthStore(t)
	ctx := context.Background()
	created, err := store.SaveRefreshJTI(ctx, 1, "session", "jti")
	if err != nil || !created {
		t.Fatalf("first SaveRefreshJTI() = %v, %v", created, err)
	}
	created, err = store.SaveRefreshJTI(ctx, 2, "other-session", "jti")
	if err != nil || created {
		t.Fatalf("second SaveRefreshJTI() = %v, %v", created, err)
	}
}
