package mux

import (
	"regexp"
	"strings"
)

var invalid = regexp.MustCompile(`[^.a-zA-Z0-9\-]`)
var prefix = regexp.MustCompile(`^[0-9\-]+`)
var suffix = regexp.MustCompile(`-$`)

// ToHostname converts a string to RFC 1123/952 hostname.
func ToHostname(str string) string {
	result := strings.ReplaceAll(str, " ", "-")
	result = invalid.ReplaceAllString(result, "")
	result = prefix.ReplaceAllString(result, "")
	result = suffix.ReplaceAllString(result, "")
	return strings.ToLower(result)
}
