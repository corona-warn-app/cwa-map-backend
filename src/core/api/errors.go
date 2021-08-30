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
			Validation: fieldError.Tag(),
		})
	}
	return ValidationErrorResponse{
		ErrorResponse: ErrorResponse{Timestamp: time.Now(), Error: "validation error"},
		Errors:        fields,
	}
}
