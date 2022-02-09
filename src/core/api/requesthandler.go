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
	"com.t-systems-mms.cwa/core/security"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io"
	"net/http"
	"time"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) (interface{}, error)

// Handle handles incoming requests and encapsulates response marshalling and error handling
func Handle(handler HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		response, err := handler(writer, request)
		if err != nil {
			WriteError(writer, request, err)
		} else if response == nil {
			logrus.WithFields(logrus.Fields{
				"path":   request.URL.Path,
				"query":  request.URL.RawQuery,
				"status": http.StatusNoContent,
			}).Debug("Handled API request")
			WriteResponse(writer, http.StatusNoContent, response)
		} else {
			logrus.WithFields(logrus.Fields{
				"path":   request.URL.Path,
				"query":  request.URL.RawQuery,
				"status": http.StatusOK,
			}).Debug("Handled API request")
			WriteResponse(writer, http.StatusOK, response)
		}

	}
}

func WriteError(writer http.ResponseWriter, request *http.Request, err error) {
	var response interface{}
	status := http.StatusInternalServerError

	switch err := err.(type) {
	case HandlerError:
		status = err.Status
		response = createHandlerErrorResponse(err)
	case validator.ValidationErrors:
		status = http.StatusBadRequest
		response = createValidationErrorResponse(err)
	default:
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
			response = ErrorResponse{Timestamp: time.Now(), Error: "record not found"}
		} else if err == security.ErrForbidden {
			response = ErrorResponse{Timestamp: time.Now(), Error: "forbidden"}
		} else {
			response = ErrorResponse{Timestamp: time.Now(), Error: "internal server error"}
		}
	}

	logrus.WithFields(logrus.Fields{
		"path":   request.URL.Path,
		"query":  request.URL.RawQuery,
		"status": status,
	}).WithError(err).Error("Error handling API request")
	WriteResponse(writer, status, response)
}

func WriteResponse(writer http.ResponseWriter, code int, body interface{}) {
	writer.Header().Add("Content-type", "application/json")
	writer.WriteHeader(code)
	if body != nil {
		responseBody, _ := json.Marshal(body)
		if _, err := writer.Write(responseBody); err != nil {
			logrus.WithError(err).Error("Error writing response")
		}
	}
}

// ParseRequestBody parses the request body, unmarshals the json into target and validates it
func ParseRequestBody(r *http.Request, validate *validator.Validate, target interface{}) error {
	if data, err := io.ReadAll(r.Body); err == nil {
		if err := json.Unmarshal(data, target); err == nil {
			if vErr := validate.Struct(target); vErr != nil {
				return vErr
			}
		} else {
			return err
		}
	} else {
		return err
	}
	return nil
}
