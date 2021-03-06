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
	"bytes"
	"com.t-systems-mms.cwa/api/model"
	"com.t-systems-mms.cwa/core/api"
	"com.t-systems-mms.cwa/core/util"
	"com.t-systems-mms.cwa/repositories"
	"com.t-systems-mms.cwa/services"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/go-playground/validator"
	"github.com/sirupsen/logrus"
	"github.com/vincent-petithory/dataurl"
	"image"
	"net/http"
)

type Operators struct {
	chi.Router
	operatorsRepository repositories.Operators
	operatorsService    services.Operators
	validate            *validator.Validate
}

func NewOperatorsAPI(operatorsRepository repositories.Operators, operatorsService services.Operators, auth *jwtauth.JWTAuth) *Operators {
	validate := validator.New()
	validate.RegisterTagNameFunc(util.JsonTagNameFunc)

	operators := &Operators{
		Router:              chi.NewRouter(),
		operatorsService:    operatorsService,
		operatorsRepository: operatorsRepository,
		validate:            validate,
	}

	operators.Get("/{operator}/logo", operators.GetOperatorLogo)
	operators.Get("/{operator}/marker", operators.GetOperatorMarker)

	operators.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth))
		r.Use(jwtauth.Authenticator)
		r.Get("/current", api.Handle(operators.GetCurrentOperator))
		r.Put("/current", api.Handle(operators.SaveCurrentOperator))
	})

	return operators
}

func (c *Operators) SaveCurrentOperator(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	operator, err := c.operatorsService.GetCurrentOperator(r.Context())
	if err != nil {
		return nil, err
	}

	request := model.OperatorDTO{}
	if err := api.ParseRequestBody(r, c.validate, &request); err != nil {
		return nil, err
	}

	operator.Name = request.Name
	operator.Email = request.Email
	operator.BugReportsReceiver = request.ReportReceiver

	if request.Logo != nil {
		data, err := dataurl.DecodeString(*request.Logo)
		if err != nil {
			logrus.WithError(err).Error("Error decoding image data")
			return nil, api.HandlerError{Status: http.StatusBadRequest, Err: "invalid image data"}
		}

		img, _, err := image.DecodeConfig(bytes.NewReader(data.Data))
		if err != nil {
			logrus.WithError(err).Error("Error reading image config")
			return nil, api.HandlerError{Status: http.StatusBadRequest, Err: "invalid image data"}
		}

		if img.Width > 100 || img.Height > 70 {
			return nil, api.HandlerError{Status: http.StatusBadRequest, Err: "image must not be larger then 100x70"}
		}
		operator.Logo = request.Logo
	}
	operator, err = c.operatorsRepository.Save(r.Context(), operator)
	if err != nil {
		return nil, err
	}
	return model.MapToOperatorDTO(&operator), nil
}

func (c *Operators) GetCurrentOperator(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	operator, err := c.operatorsService.GetCurrentOperator(r.Context())
	if err != nil {
		return nil, err
	}

	return model.MapToOperatorDTO(&operator), nil
}

func (c *Operators) GetOperatorMarker(w http.ResponseWriter, r *http.Request) {
	operatorId := chi.URLParam(r, "operator")
	operator, err := c.operatorsRepository.FindById(r.Context(), operatorId)
	if err != nil {
		logrus.WithError(err).Error("Error getting operator")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if operator.MarkerIcon != nil {
		dataUrl, err := dataurl.DecodeString(*operator.MarkerIcon)
		if err != nil {
			logrus.WithError(err).Error("Error parsing marker icon")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", dataUrl.ContentType())
		_, _ = w.Write(dataUrl.Data)
	}
	w.WriteHeader(http.StatusNotFound)
}

func (c *Operators) GetOperatorLogo(w http.ResponseWriter, r *http.Request) {
	operatorId := chi.URLParam(r, "operator")
	operator, err := c.operatorsRepository.FindById(r.Context(), operatorId)
	if err != nil {
		logrus.WithError(err).Error("Error getting operator")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if operator.Logo != nil {
		dataUrl, err := dataurl.DecodeString(*operator.Logo)
		if err != nil {
			logrus.WithError(err).Error("Error parsing operator logo")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", dataUrl.ContentType())
		_, _ = w.Write(dataUrl.Data)
	}
	w.WriteHeader(http.StatusNotFound)
}
