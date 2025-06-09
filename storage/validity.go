package storage

import (
	"errors"
	"strconv"
	"time"
)

func parseValidity(n time.Time, validity string) (time.Time, error) {
	m := validityRegExp.FindStringSubmatch(validity)
	if len(m) != 3 {
		return time.Time{}, errors.New("invalid validity specified")
	}
	v, err := strconv.Atoi(m[1])
	if err != nil {
		return time.Time{}, err
	}
	switch m[2] {
	case "d":
		return n.Add(time.Duration(v) * 24 * time.Hour), nil
	case "y":
		return time.Date(
			n.Year()+v,
			n.Month(),
			n.Day(),
			n.Hour(),
			n.Minute(),
			n.Second(),
			n.Nanosecond(),
			n.Location(),
		), nil
	default:
		return time.Time{}, errors.New("invalid time unit specified")
	}
}
