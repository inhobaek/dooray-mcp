package utils

import (
	"fmt"
	"time"
)

const iso8601 = "2006-01-02T15:04:05-07:00"
const iso8601DateOnly = "2006-01-02-07:00"

func ConvertISO8601ToTime(t string) (time.Time, error) {
	parsedTime, err := time.Parse(iso8601, t)
	if err == nil {
		return parsedTime, nil
	}
	parsedTime, err2 := time.Parse(iso8601DateOnly, t)
	if err2 == nil {
		return parsedTime, nil
	}
	return time.Time{}, fmt.Errorf("failed to parse time: %w", err)
}

func IsDateOnlyFormat(t string) bool {
	_, err := time.Parse(iso8601DateOnly, t)
	return err == nil
}

func FormatTimeToISO8601(t time.Time) string {
	return t.Format(iso8601)
}

func FormatDateOnly(t time.Time) string {
	return t.Format(iso8601DateOnly)
}

