package tools

import (
	"fmt"
	"time"
)

func getCurrentTime(timezone string) string {
	if timezone == "" {
		return time.Now().Format("2006-01-02 15:04:05 MST")
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Sprintf("error: unknown timezone %q", timezone)
	}
	return time.Now().In(loc).Format("2006-01-02 15:04:05 MST")
}
