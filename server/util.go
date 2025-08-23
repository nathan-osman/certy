package server

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
)

func formatDate(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	v, ok := in.Interface().(time.Time)
	if !ok {
		return nil, &pongo2.Error{
			Sender:    "filter:formatDate",
			OrigError: errors.New("time.Time required"),
		}
	}
	return pongo2.AsValue(
		v.Format("2006-01-02 3:04 PM"),
	), nil
}

func formatBytes(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	v, ok := in.Interface().([]uint8)
	if !ok {
		return nil, &pongo2.Error{
			Sender:    "filter:formatBytes",
			OrigError: errors.New("[]uint8 required"),
		}
	}
	values := []string{}
	for _, b := range v {
		values = append(values, fmt.Sprintf("%02X", b))
	}
	return pongo2.AsValue(
		strings.Join(values, ":"),
	), nil
}
