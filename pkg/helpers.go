package pkg

import (
	"encoding/base64"
	"log/slog"
	"os"
	"unsafe"
)

const defaultLogLevel = 0 // info

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

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
}
