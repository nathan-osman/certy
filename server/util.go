package server

import (
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
	"github.com/nathan-osman/certy/storage"
)

const (
	durDay  = 24 * time.Hour
	durYear = 365 * durDay
)

// When capturing the stack, we need to skip five frames:
// - runtime.Callers() itself
// - captureStack()
// - errorHandler()
// - Gin's internal recovery function
// - the call to panic()

func captureStack() string {
	var (
		pcs    = make([]uintptr, 64)
		n      = runtime.Callers(5, pcs)
		frames = runtime.CallersFrames(pcs[:n])
		lines  []string
	)
	for {
		f, more := frames.Next()
		lines = append(
			lines,
			fmt.Sprintf(
				"%s:%d\n\t%s",
				f.File,
				f.Line,
				f.Function,
			),
		)
		if !more {
			break
		}
	}
	return strings.Join(lines, "\n")
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

func formatDuration(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	v, ok := in.Interface().(time.Duration)
	if !ok {
		return nil, &pongo2.Error{
			Sender:    "filter:formatDuration",
			OrigError: errors.New("time.Duration required"),
		}
	}
	var vStr string
	switch {
	case v < durDay:
		vStr = fmt.Sprintf("%dh", v/time.Hour)
	case v < durYear:
		vStr = fmt.Sprintf("%dd", v/durDay)
	default:
		vStr = fmt.Sprintf("%dy", v/durYear)
	}
	return pongo2.AsValue(
		vStr,
	), nil
}

func ifPresent(v []string) string {
	if len(v) > 0 {
		return v[0]
	}
	return ""
}

func combineAddress(n pkix.Name) string {
	parts := []string{}
	for _, v := range [][]string{
		n.StreetAddress,
		n.Locality,
		n.Province,
		n.Country,
		n.PostalCode,
	} {
		if len(v) > 0 {
			parts = append(parts, v[0])
		}
	}
	return strings.Join(parts, ", ")
}

func downloadCert(
	c *gin.Context,
	mime string,
	b []byte,
	cert *storage.Certificate,
	suffix, extension string,
) {
	v := strings.ReplaceAll(
		cert.X509.Subject.CommonName,
		" ",
		"_",
	)
	c.Header(
		"Content-Disposition",
		fmt.Sprintf(
			`attachment; filename="%s%s.%s"`,
			v,
			suffix,
			extension,
		),
	)
	c.Header("Content-Length", strconv.Itoa(len(b)))
	c.Data(http.StatusOK, mime, b)
}
