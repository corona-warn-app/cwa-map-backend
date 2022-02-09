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

package api

import (
	"github.com/go-playground/validator"
	"time"
)

type HandlerError struct {
	Status int
	Err    string
}

func (h HandlerError) Error() string {
	return h.Err
}

type ErrorResponse struct {
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"message"`
}

type ValidationFieldError struct {
	Field      string `json:"field"`
	Validation string `json:"validation"`
}

type ValidationErrorResponse struct {
	ErrorResponse
	Errors []ValidationFieldError `json:"errors"`
}

func createHandlerErrorResponse(err error) ErrorResponse {
	return ErrorResponse{Timestamp: time.Now(), Error: err.Error()}
}

func createValidationErrorResponse(errors validator.ValidationErrors) ValidationErrorResponse {
	fields := make([]ValidationFieldError, 0)
	for _, fieldError := range errors {
		fields = append(fields, ValidationFieldError{
			Field:      fieldError.Field(),
			Validation: fieldError.Tag() + "=" + fieldError.Param(),
		})
	}
	return ValidationErrorResponse{
		ErrorResponse: ErrorResponse{Timestamp: time.Now(), Error: "validation error"},
		Errors:        fields,
	}
}
