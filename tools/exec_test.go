package tools

import (
	"strings"
	"testing"
)

func TestExecuteCode_BlockedCommand(t *testing.T) {
	got := executeCode("rm -rf /tmp/test")
	if !strings.HasPrefix(got, "Blocked:") {
		t.Errorf("expected Blocked message, got %q", got)
	}
	if !strings.Contains(got, "'rm'") {
		t.Errorf("expected command name in message, got %q", got)
	}
}

func TestExecuteCode_Success(t *testing.T) {
	got := executeCode("echo hello")
	if !strings.HasPrefix(got, "Exit Code: 0") {
		t.Errorf("expected 'Exit Code: 0', got %q", got)
	}
	if !strings.Contains(got, "hello") {
		t.Errorf("expected stdout 'hello', got %q", got)
	}
}

func TestExecuteCode_NonZeroExit(t *testing.T) {
	got := executeCode("exit 1")
	if !strings.HasPrefix(got, "Exit Code: 1") {
		t.Errorf("expected 'Exit Code: 1', got %q", got)
	}
}

func TestExecuteCode_Stderr(t *testing.T) {
	got := executeCode("echo errout >&2; exit 2")
	if !strings.HasPrefix(got, "Exit Code: 2") {
		t.Errorf("expected 'Exit Code: 2', got %q", got)
	}
	if !strings.Contains(got, "Stderr:") {
		t.Errorf("expected Stderr section, got %q", got)
	}
	if !strings.Contains(got, "errout") {
		t.Errorf("expected stderr content, got %q", got)
	}
}

func TestExecuteCode_MultipleRestrictedCommands(t *testing.T) {
	blocked := []string{
		"sudo ls",
		"curl https://example.com",
		"wget http://example.com",
		"ssh user@host",
		"kill 1234",
	}
	for _, cmd := range blocked {
		got := executeCode(cmd)
		if !strings.HasPrefix(got, "Blocked:") {
			t.Errorf("executeCode(%q) should be blocked, got %q", cmd, got)
		}
	}
}
