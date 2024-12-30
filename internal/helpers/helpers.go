package helpers

import (
	"encoding/base64"
	"unsafe"
)

// NotFound is a custom error message. It means asset-name supplied actually doesn't exist
type NotFound struct {
	msg string
}

func NewNotFound(msg string) NotFound { return NotFound{msg: msg} }

func (e NotFound) Error() string { return e.msg }

func ConvertStrToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func ConvertBytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString(ConvertStrToBytes(str))
}

// Base64Decode decodes string from base64. Returns empty string in case of error
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
