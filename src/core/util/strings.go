/*
 *   Corona-Warn-App / cwa-map-backend
 *
 *   (C) 2020, T-Systems International GmbH
 *
 *   Deutsche Telekom AG and all other contributors /
 *   copyright owners license this file to you under the Apache
 *   License, Version 2.0 (the "License"); you may not use this
 *   file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing,
 *   software distributed under the License is distributed on an
 *   "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *   KIND, either express or implied.  See the License for the
 *   specific language governing permissions and limitations
 *   under the License.
 */

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
