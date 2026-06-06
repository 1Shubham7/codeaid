package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("error reading file: %v", err)
	}
	return string(data)
}

func listDirectory(path string) string {
	if path == "" {
		path = "."
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Sprintf("error reading directory: %v", err)
	}

	var dirs, files []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, "  "+e.Name()+"/")
		} else {
			files = append(files, "  "+e.Name())
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Directory: %s\n\n", path))
	if len(dirs) > 0 {
		sb.WriteString("Directories:\n")
		for _, d := range dirs {
			sb.WriteString(d + "\n")
		}
	}
	if len(files) > 0 {
		if len(dirs) > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("Files:\n")
		for _, f := range files {
			sb.WriteString(f + "\n")
		}
	}
	if len(dirs) == 0 && len(files) == 0 {
		sb.WriteString("(empty)")
	}
	return sb.String()
}

func writeFile(path, content string) string {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Sprintf("error creating directories: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Sprintf("error writing file: %v", err)
	}
	return fmt.Sprintf("successfully wrote %d bytes to %s", len(content), path)
}
