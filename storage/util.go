package storage

import (
	"os"
)

func ifProvided(v string) []string {
	if v == "" {
		return []string{}
	}
	return []string{v}
}

func fileExists(f string) (bool, error) {
	_, err := os.Stat(f)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
