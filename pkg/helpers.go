package pkg

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"unsafe"
)

const defaultLogLevel = 0 // info

// ErrNotFound is a sentinel error to indicate resource not found.
var ErrNotFound = errors.New("resource not found")

// NewNotFoundError creates a formatted not-found error.
func NewNotFoundError(resource string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, resource)
}

func ConvertStrToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func ConvertBytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString(ConvertStrToBytes(str))
}

// Base64Decode decodes string from base64. Returns empty string in case of an error
func Base64Decode(str string) string {
	if str == "" {
		return ""
	}
	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return ""
	}
	return ConvertBytesToString(b)
}

func DefaultLogger() *slog.Logger {
	var logLevel = new(slog.LevelVar)
	logLevel.Set(defaultLogLevel)

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
}
