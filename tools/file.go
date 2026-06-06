package tools

import (
	"fmt"
	"os"
)

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("error reading file: %v", err)
	}

	return string(data)
}
