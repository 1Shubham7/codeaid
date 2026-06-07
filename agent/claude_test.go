package agent

import (
	"encoding/json"
	"testing"
)

func TestToolSummary_ExecuteCode_Success(t *testing.T) {
	result := "Exit Code: 0\n\nStdout:\nhello\n"
	tc := toolSummary("execute_code", json.RawMessage(`{"command":"echo hello"}`), result)

	if tc.IsError {
		t.Error("IsError should be false for exit code 0")
	}
	if tc.Display != "ran: echo hello" {
		t.Errorf("Display = %q, want %q", tc.Display, "ran: echo hello")
	}
	if tc.Output != result {
		t.Errorf("Output = %q, want %q", tc.Output, result)
	}
}

func TestToolSummary_ExecuteCode_Failure(t *testing.T) {
	result := "Exit Code: 1\n\nStderr:\ncommand not found\n"
	tc := toolSummary("execute_code", json.RawMessage(`{"command":"badcmd"}`), result)

	if !tc.IsError {
		t.Error("IsError should be true for non-zero exit code")
	}
	if tc.Display != "ran: badcmd" {
		t.Errorf("Display = %q, want %q", tc.Display, "ran: badcmd")
	}
}

func TestToolSummary_ExecuteCode_Blocked(t *testing.T) {
	result := "Blocked: 'rm' is a restricted command and was not executed."
	tc := toolSummary("execute_code", json.RawMessage(`{"command":"rm -rf /"}`), result)

	if !tc.IsError {
		t.Error("IsError should be true for blocked command (no 'Exit Code: 0' prefix)")
	}
}

func TestToolSummary_ReadFile(t *testing.T) {
	tc := toolSummary("read_file", json.RawMessage(`{"path":"main.go"}`), "package main")

	if tc.Display != "read file: main.go" {
		t.Errorf("Display = %q, want %q", tc.Display, "read file: main.go")
	}
	if tc.Output != "" {
		t.Errorf("Output should be empty, got %q", tc.Output)
	}
	if tc.IsError {
		t.Error("IsError should be false")
	}
}

func TestToolSummary_WriteFile(t *testing.T) {
	tc := toolSummary("write_file", json.RawMessage(`{"path":"out.txt","content":"hi"}`), "successfully wrote 2 bytes to out.txt")

	if tc.Display != "wrote file: out.txt" {
		t.Errorf("Display = %q, want %q", tc.Display, "wrote file: out.txt")
	}
}

func TestToolSummary_ListDirectory_EmptyPath(t *testing.T) {
	tc := toolSummary("list_directory", json.RawMessage(`{}`), "Directory: .\n")

	if tc.Display != "listed directory: ." {
		t.Errorf("Display = %q, want %q", tc.Display, "listed directory: .")
	}
}

func TestToolSummary_ListDirectory_WithPath(t *testing.T) {
	tc := toolSummary("list_directory", json.RawMessage(`{"path":"./cmd"}`), "Directory: ./cmd\n")

	if tc.Display != "listed directory: ./cmd" {
		t.Errorf("Display = %q, want %q", tc.Display, "listed directory: ./cmd")
	}
}

func TestToolSummary_GetCurrentTime(t *testing.T) {
	tc := toolSummary("get_current_time", json.RawMessage(`{}`), "2024-01-01 12:00:00 UTC")

	if tc.Display != "fetched current time" {
		t.Errorf("Display = %q, want %q", tc.Display, "fetched current time")
	}
	if tc.Output != "" || tc.IsError {
		t.Errorf("expected empty Output and IsError=false")
	}
}

func TestToolSummary_UnknownTool(t *testing.T) {
	tc := toolSummary("some_future_tool", json.RawMessage(`{}`), "result")

	if tc.Display != "called tool: some_future_tool" {
		t.Errorf("Display = %q, want %q", tc.Display, "called tool: some_future_tool")
	}
}
