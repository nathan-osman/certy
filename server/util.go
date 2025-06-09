package server

import (
	"errors"
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
		v.Format("January 2, 2006 3:04:05 PM"),
	), nil
}
