package storage

import (
	"os"
	"path"
	"strconv"
)

const (
	filenameSerial = "serial"
)

func (s *Storage) allocNextSerial(dir string) (int64, error) {
	var (
		filename       = path.Join(dir, filenameSerial)
		serial   int64 = 1
	)
	b, err := os.ReadFile(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return 0, err
		}
	} else {
		v, err := strconv.ParseInt(string(b), 10, 64)
		if err != nil {
			return 0, nil
		}
		serial = v + 1
	}
	if err := os.WriteFile(
		filename,
		[]byte(strconv.FormatInt(serial, 10)),
		0600,
	); err != nil {
		return 0, err
	}
	return serial, nil
}
