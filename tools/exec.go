package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/1shubham7/codeaid/logger"
)

func executeCode(command string) string {
	logger.L.Info("execute_code", "command", command)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	if wd, err := os.Getwd(); err == nil {
		cmd.Dir = wd
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	if exitCode == 0 {
		logger.L.Info("execute_code success", "command", command, "stdout_bytes", stdout.Len())
	} else {
		logger.L.Warn("execute_code failed", "command", command, "exit_code", exitCode, "stderr", stderr.String())
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Exit Code: %d\n", exitCode))
	if stdout.Len() > 0 {
		sb.WriteString("\nStdout:\n")
		sb.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		sb.WriteString("\nStderr:\n")
		sb.WriteString(stderr.String())
	}
	return sb.String()
}
