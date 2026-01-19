package dates

import (
	"fmt"
	"strings"
	"time"
)

const Layout = "01-2006"

func String2Date(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	t, err := time.Parse(Layout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format")
	}

	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil // Normalize to single day and time, we only care about month and year
}
