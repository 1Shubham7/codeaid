package tools

import (
	"os"
	"strings"
	"testing"

	"github.com/1shubham7/codeaid/logger"
)

func TestMain(m *testing.M) {
	logger.InitForTesting()
	SetRestrictedCommands(DefaultRestrictedCommands)
	SetMaxFileSizeKB(100)
	os.Exit(m.Run())
}

func TestIsRestricted(t *testing.T) {
	tests := []struct {
		cmd     string
		want    bool
		wantBin string
	}{
		{"rm -rf /", true, "rm"},
		{"sudo apt install vim", true, "sudo"},
		{"curl https://example.com", true, "curl"},
		{"wget http://example.com", true, "wget"},
		{"ssh user@host", true, "ssh"},
		{"echo hello", false, ""},
		{"go build .", false, ""},
		{"python3 main.py", false, ""},
		{"rm", true, "rm"},
		{"", false, ""},
	}
	for _, tt := range tests {
		bin, got := isRestricted(tt.cmd)
		if got != tt.want {
			t.Errorf("isRestricted(%q) blocked = %v, want %v", tt.cmd, got, tt.want)
		}
		if tt.want && bin != tt.wantBin {
			t.Errorf("isRestricted(%q) bin = %q, want %q", tt.cmd, bin, tt.wantBin)
		}
	}
}

func TestDispatch_UnknownTool(t *testing.T) {
	result := Dispatch("nonexistent_tool", nil)
	if !strings.HasPrefix(result, "unknown tool:") {
		t.Errorf("expected 'unknown tool:' prefix, got %q", result)
	}
	if !strings.Contains(result, "nonexistent_tool") {
		t.Errorf("expected tool name in result, got %q", result)
	}
}
