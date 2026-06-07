package tools

import (
	"strings"
	"testing"
)

func TestGetCurrentTime_Default(t *testing.T) {
	got := getCurrentTime("")
	// Expect "YYYY-MM-DD HH:MM:SS TZ" — at least 19 chars.
	if len(got) < 19 {
		t.Errorf("unexpected time format: %q", got)
	}
}

func TestGetCurrentTime_ValidTimezone(t *testing.T) {
	got := getCurrentTime("UTC")
	if !strings.HasSuffix(got, "UTC") {
		t.Errorf("expected UTC suffix, got %q", got)
	}
}

func TestGetCurrentTime_AnotherTimezone(t *testing.T) {
	got := getCurrentTime("America/New_York")
	if strings.HasPrefix(got, "error:") {
		t.Errorf("unexpected error for valid timezone: %q", got)
	}
	if len(got) < 19 {
		t.Errorf("unexpected time format: %q", got)
	}
}

func TestGetCurrentTime_InvalidTimezone(t *testing.T) {
	got := getCurrentTime("Not/A/Timezone")
	if !strings.HasPrefix(got, "error:") {
		t.Errorf("expected error message, got %q", got)
	}
	if !strings.Contains(got, "Not/A/Timezone") {
		t.Errorf("expected timezone name in error, got %q", got)
	}
}
