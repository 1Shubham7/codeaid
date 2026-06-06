package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func executeCode(command string) string {
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
