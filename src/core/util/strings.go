package util

import (
	"strings"
	"time"
)

func IsNilOrEmpty(value *string) bool {
	return value == nil || strings.TrimSpace(*value) == ""
}

func IsNotNilOrEmpty(value *string) bool {
	return !IsNilOrEmpty(value)
}

// PtrToString returns an empty string for nil or the string value
func PtrToString(value *string, nilValue string) string {
	if value == nil {
		return nilValue
	}
	return *value
}

func StringAsPtr(value string) *string {
	return &value
}

func BoolToString(value *bool, nilValue string) string {
	if value == nil {
		return nilValue
	} else if *value {
		return "true"
	} else {
		return "false"
	}
}

func TimeToString(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func ArrayContainsOne(arr []string, search ...string) bool {
	for _, v := range arr {
		for _, s := range search {
			if v == s {
				return true
			}
		}
	}
	return false
}
