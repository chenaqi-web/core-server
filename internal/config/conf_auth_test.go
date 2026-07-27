package config

import (
	"testing"
	"time"
)

func TestAuthConfigDurations(t *testing.T) {
	cfg := AuthConfig{AccessExpire: "20m", RefreshExpire: "168h"}

	accessDuration, err := cfg.AccessDuration()
	if err != nil {
		t.Fatalf("AccessDuration() error = %v", err)
	}
	if accessDuration != 20*time.Minute {
		t.Fatalf("AccessDuration() = %v, want %v", accessDuration, 20*time.Minute)
	}

	refreshDuration, err := cfg.RefreshDuration()
	if err != nil {
		t.Fatalf("RefreshDuration() error = %v", err)
	}
	if refreshDuration != 7*24*time.Hour {
		t.Fatalf("RefreshDuration() = %v, want %v", refreshDuration, 7*24*time.Hour)
	}
}

func TestAuthConfigDurationRejectsInvalidValue(t *testing.T) {
	if _, err := (AuthConfig{AccessExpire: "0s"}).AccessDuration(); err == nil {
		t.Fatal("AccessDuration() expected error for non-positive duration")
	}
}
