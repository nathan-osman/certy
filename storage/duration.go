package storage

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

const (
	durDay   = 24 * time.Hour
	durWeek  = 7 * durDay
	durMonth = 30 * durDay
	durYear  = 365 * durDay
)

var (
	validityRegExp = regexp.MustCompile(`(\d+)(\w+)`)

	errInvalidDuration = errors.New("invalid duration specified")
	errInvalidUnit     = errors.New("invalid unit specified")
)

func parseDuration(v string) (time.Duration, error) {
	m := validityRegExp.FindStringSubmatch(v)
	if len(m) != 3 {
		return 0, errInvalidDuration
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, err
	}
	switch m[2] {
	case "h":
		return time.Duration(n) * time.Hour, nil
	case "d":
		return time.Duration(n) * durDay, nil
	case "w":
		return time.Duration(n) * durWeek, nil
	case "m":
		return time.Duration(n) * durMonth, nil
	case "y":
		return time.Duration(n) * durYear, nil
	default:
		return 0, errInvalidUnit
	}
}
