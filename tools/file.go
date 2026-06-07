package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/1shubham7/codeaid/logger"
)

func readFile(path string) string {
	logger.L.Info("read_file", "path", path)

	info, err := os.Stat(path)
	if err != nil {
		logger.L.Error("read_file stat failed", "path", path, "err", err)
		return fmt.Sprintf("error accessing file: %v", err)
	}
	if info.Size() > maxFileSizeBytes {
		limitKB := maxFileSizeBytes / 1024
		logger.L.Warn("read_file blocked: file too large", "path", path, "size_bytes", info.Size(), "limit_kb", limitKB)
		return fmt.Sprintf("Blocked: file '%s' is %d KB which exceeds the %d KB limit.", path, info.Size()/1024, limitKB)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logger.L.Error("read_file failed", "path", path, "err", err)
		return fmt.Sprintf("error reading file: %v", err)
	}
	logger.L.Info("read_file success", "path", path, "bytes", len(data))
	return string(data)
}

func listDirectory(path string) string {
	if path == "" {
		path = "."
	}
	logger.L.Info("list_directory", "path", path)
	entries, err := os.ReadDir(path)
	if err != nil {
		logger.L.Error("list_directory failed", "path", path, "err", err)
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

	logger.L.Info("list_directory success", "path", path, "dirs", len(dirs), "files", len(files))

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
	logger.L.Info("write_file", "path", path, "bytes", len(content))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		logger.L.Error("write_file mkdir failed", "path", path, "err", err)
		return fmt.Sprintf("error creating directories: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		logger.L.Error("write_file failed", "path", path, "err", err)
		return fmt.Sprintf("error writing file: %v", err)
	}
	logger.L.Info("write_file success", "path", path, "bytes", len(content))
	return fmt.Sprintf("successfully wrote %d bytes to %s", len(content), path)
}
