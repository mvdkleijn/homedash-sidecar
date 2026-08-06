package config

import (
	"io"
	"log/slog"
	"testing"
	"time"
)

func init() {
	SetLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestLoadInvalidIntervalUsesDefault(t *testing.T) {
	t.Setenv("HOMEDASH_SERVER", "http://server.test")
	t.Setenv("HOMEDASH_INTERVAL", "not-a-duration")

	cfg := Load()

	if cfg.Interval != 10*time.Minute {
		t.Errorf("expected default interval 10m, got %v", cfg.Interval)
	}
}

func TestLoadValidInterval(t *testing.T) {
	t.Setenv("HOMEDASH_SERVER", "http://server.test")
	t.Setenv("HOMEDASH_INTERVAL", "5m")

	cfg := Load()

	if cfg.Interval != 5*time.Minute {
		t.Errorf("expected interval 5m, got %v", cfg.Interval)
	}
}

func TestLoadDefaultIntervalWhenUnset(t *testing.T) {
	t.Setenv("HOMEDASH_SERVER", "http://server.test")
	t.Setenv("HOMEDASH_INTERVAL", "")

	cfg := Load()

	if cfg.Interval != 10*time.Minute {
		t.Errorf("expected default interval 10m, got %v", cfg.Interval)
	}
}
