package logger

import (
	"log/slog"
	"os"
	"path/filepath"
)

// L is the package-level logger. Call Init() once at startup before using it.
var L *slog.Logger

// Init creates ~/.codeaid/logs/codeaid.log and sets up L. Must be called before
// any logging. Falls back to stderr if the file cannot be opened.
func Init() {
	home, err := os.UserHomeDir()
	if err != nil {
		L = slog.New(slog.NewJSONHandler(os.Stderr, nil))
		return
	}

	logDir := filepath.Join(home, ".codeaid", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		L = slog.New(slog.NewJSONHandler(os.Stderr, nil))
		return
	}

	f, err := os.OpenFile(
		filepath.Join(logDir, "codeaid.log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		L = slog.New(slog.NewJSONHandler(os.Stderr, nil))
		return
	}

	L = slog.New(slog.NewJSONHandler(f, nil))
}
