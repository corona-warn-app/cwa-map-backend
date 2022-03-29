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
	"com.t-systems-mms.cwa/api/model"
	"com.t-systems-mms.cwa/core/api"
	"com.t-systems-mms.cwa/core/security"
	"com.t-systems-mms.cwa/core/util"
	"com.t-systems-mms.cwa/domain"
	"com.t-systems-mms.cwa/external/geocoding"
	"com.t-systems-mms.cwa/repositories"
	"com.t-systems-mms.cwa/services"
	"context"
	"encoding/csv"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/go-playground/validator"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
)

var (
	ErrInvalidAddress    = api.HandlerError{Status: http.StatusBadRequest, Err: "invalid address"}
	ErrInvalidParameters = api.HandlerError{Status: http.StatusBadRequest, Err: "invalid parameters"}
)

var (
	findCentersRequestsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cwa_map_find_centers_request_count",
		Help: "The total count of find centers requests",
	})

	deliveredCentersCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cwa_map_delivered_centers_count",
		Help: "The total count of centers delivered to the users",
	})

	emptyCentersCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cwa_map_empty_centers_count",
		Help: "The total count of requests returning no centers",
	})

	geocodeRequestsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cwa_map_geocode_request_count",
		Help: "The total count of geocode requests",
	})
)

type Centers struct {
	chi.Router
	centersService    services.Centers
	geocoder          geocoding.Geocoder
	operatorsService  services.Operators
	centersRepository repositories.Centers
	bugReportsService services.BugReports
	validate          *validator.Validate
}

func NewCentersAPI(centersService services.Centers, centersRepository repositories.Centers,
	bugReportsService services.BugReports,
	operatorsService services.Operators, geocoder geocoding.Geocoder, auth *jwtauth.JWTAuth) *Centers {
	validate := validator.New()
	validate.RegisterTagNameFunc(util.JsonTagNameFunc)

	centers := &Centers{
		Router:            chi.NewRouter(),
		centersService:    centersService,
		centersRepository: centersRepository,
		operatorsService:  operatorsService,
		geocoder:          geocoder,
		bugReportsService: bugReportsService,
		validate:          validate,
	}

	// public endpoints
	centers.Get("/", api.Handle(centers.findCenters))
	centers.Get("/bounds", api.Handle(centers.geocode))
	centers.Post("/{uuid}/report", api.Handle(centers.createBugReport))

	centers.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth))
		r.Use(jwtauth.Authenticator)

		r.Get("/all", api.Handle(centers.getAllCenters))
		r.Post("/csv", api.Handle(centers.prepareCSVImport))
		r.Post("/", api.Handle(centers.importCenters))
		r.Put("/{uuid}", api.Handle(centers.updateCenter))

		// get centers
		r.Get("/reference/{reference}", api.Handle(centers.getCenterByReferenceLegacy))
		r.Get("/ref/{reference}", api.Handle(centers.getCenterByReference))
		r.Get("/{uuid}", api.Handle(centers.getCenterByUUID))

		// delete centers
		r.Delete("/{uuid}", api.Handle(centers.deleteCenterByUUID))
		r.Delete("/reference/{reference}", api.Handle(centers.deleteCenterByReference))

		r.Route("/admin", func(r chi.Router) {
			r.Use(api.RequireRole(security.RoleAdmin))
			r.Get("/csv", centers.exportCentersAsCSV)
			r.Post("/geocode", api.Handle(centers.geocodeAllCenters))
		})
	})
	return centers
}

func (c *Centers) geocode(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	geocodeRequestsCounter.Inc()
	if address, hasAddress := r.URL.Query()["address"]; hasAddress {
		if util.IsNilOrEmpty(&address[0]) {
			return nil, ErrInvalidAddress
		}

		if result, err := c.geocoder.GetCoordinates(r.Context(), address[0]); err == nil {
			return model.GeocodeResultDTO{}.MapFromModel(&result), nil
		} else {
			if err == geocoding.ErrNoResult || err == geocoding.ErrTooManyResults {
				return nil, api.HandlerError{Status: http.StatusBadRequest, Err: err.Error()}
			}
			return nil, err
		}
	}
	return nil, ErrInvalidParameters
}

func (c *Centers) findCenters(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	findCentersRequestsCounter.Inc()
	if bounds, hasBounds, err := c.getBoundsParameter(r); hasBounds && err == nil {
		searchParameters := c.getSearchParameters(r)
		centers, err := c.centersRepository.FindByBounds(r.Context(), bounds, searchParameters, 200)
		if err != nil {
			logrus.WithError(err).Error("Error getting centers")
			return nil, err
		}

		centersCount := len(centers)
		if centersCount == 0 {
			emptyCentersCounter.Inc()
		}
		deliveredCentersCounter.Add(float64(centersCount))

		return model.FindCentersResult{
			Centers: model.MapToCenterSummaries(centers),
		}, nil
	} else if !hasBounds {
		return nil, ErrInvalidParameters
	} else {
		return nil, err
	}
}

func (c *Centers) prepareCSVImport(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	parser := &services.CsvParser{}
	result, err := parser.Parse(r.Body)
	if parseError, isParseError := err.(*csv.ParseError); isParseError {
		return nil, api.HandlerError{
			Status: http.StatusBadRequest,
			Err:    parseError.Unwrap().Error(),
		}
	} else if err != nil {
		return nil, err
	}

	return model.MapToImportCenterResultDTOs(result), nil
}

func (c *Centers) getAllCenters(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	operator, err := c.operatorsService.GetCurrentOperator(r.Context())
	if err != nil {
		return nil, err
	}

	searchString := ""
	if value, ok := r.URL.Query()["search"]; ok {
		searchString = value[0]
	}

	centers, err := c.centersRepository.FindByOperator(r.Context(), operator.UUID, searchString, repositories.ParsePageRequest(r))
	if err != nil {
		return nil, err
	}
	return model.PageCenterDTO{
		PagedResult: api.PagedResult{Count: centers.Count},
		Result:      model.MapToCenterDTOs(centers.Result),
	}, nil
}

func (c *Centers) geocodeAllCenters(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	centers, err := c.centersRepository.FindAll()
	if err != nil {
		return nil, err
	}
	go c.centersService.PerformGeocoding(context.Background(), centers)
	return nil, nil
}

// exportCentersAsCSV exports all centers as csv file
func (c *Centers) exportCentersAsCSV(w http.ResponseWriter, r *http.Request) {
	centers, err := c.centersRepository.FindAll()
	if err != nil {
		api.WriteError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		logrus.WithError(err).Error("Error writing BOM")
		return
	}

	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ';'

	if err := csvWriter.Write([]string{"partner_subject", "partner_uuid", "partner_name", "partner_number", "user_reference", "operator_name", "lab_id", "center_uuid", "center_name", "email", "address", "zip", "region", "dcc", "enter_date", "leave_date", "testkinds", "appointment", "longitude", "latitude", "message", "last_update", "visible"}); err != nil {
		logrus.WithError(err).Error("Error writing response")
		return
	}

	for _, center := range centers {
		if err := csvWriter.Write([]string{
			util.PtrToString(center.Operator.Subject, "nil"),
			center.Operator.UUID,
			center.Operator.Name,
			util.PtrToString(center.Operator.OperatorNumber, ""),
			util.PtrToString(center.UserReference, ""),
			util.PtrToString(center.OperatorName, ""),
			util.PtrToString(center.LabId, ""),
			center.UUID,
			center.Name,
			util.PtrToString(center.Email, ""),
			center.Address,
			util.PtrToString(center.Zip, ""),
			geocoding.GetRegionTranslation(center.Region),
			util.BoolToString(center.DCC, "false"),
			util.TimeToString(center.EnterDate),
			util.TimeToString(center.LeaveDate),
			strings.Join(center.TestKinds, ","),
			util.PtrToString((*string)(center.Appointment), ""),
			strconv.FormatFloat(center.Longitude, 'f', 10, 64),
			strconv.FormatFloat(center.Latitude, 'f', 10, 64),
			util.PtrToString(center.Message, ""),
			util.TimeToString(center.LastUpdate),
			util.BoolToString(center.Visible, "false"),
		}); err != nil {
			logrus.WithError(err).Error("Error writing response")
			return
		}
	}
	csvWriter.Flush()
}

func (c *Centers) importCenters(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	var importData model.ImportCenterRequest
	if err := api.ParseRequestBody(r, c.validate, &importData); err != nil {
		return nil, err
	}

	centers := make([]domain.Center, len(importData.Centers))
	for i, center := range importData.Centers {
		centers[i] = *center.MapToDomain()
	}

	result, err := c.centersService.ImportCenters(r.Context(), centers, importData.DeleteAll)
	if err != nil {
		return nil, err
	}
	return model.MapToCenterDTOs(result), nil
}

func (c *Centers) updateCenter(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	centerUUID := chi.URLParam(r, "uuid")
	logrus.WithField("uuid", centerUUID).Trace("updateCenter")

	operator, err := c.operatorsService.GetCurrentOperator(r.Context())
	if err != nil {
		return nil, err
	}

	center, err := c.centersRepository.FindByUUID(r.Context(), centerUUID)
	if err != nil {
		return nil, err
	}

	if center.OperatorUUID != operator.UUID && !security.HasRole(r.Context(), security.RoleAdmin) {
		return nil, gorm.ErrRecordNotFound
	}

	var editCenterDTO model.EditCenterDTO
	if err := api.ParseRequestBody(r, c.validate, &editCenterDTO); err != nil {
		return nil, err
	}

	editCenterDTO.CopyToDomain(&center)
	if err = c.centersService.Save(r.Context(), &center, true); err != nil {
		return nil, err
	}
	return model.CenterDTO{}.MapFromDomain(&center), nil
}

// getCenterByUUID returns the center with the given uuid.
// If the center does not belong to the currently authenticated operator, this method will return an error
func (c *Centers) getCenterByUUID(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	centerUUID := chi.URLParam(r, "uuid")
	operator, err := c.operatorsService.GetCurrentOperator(r.Context())
	if err != nil {
		return nil, err
	}

	center, err := c.centersRepository.FindByUUID(r.Context(), centerUUID)
	if err != nil {
		return nil, err
	}

	if center.OperatorUUID != operator.UUID && !security.HasRole(r.Context(), security.RoleAdmin) {
		return nil, gorm.ErrRecordNotFound
	}

	return model.CenterDTO{}.MapFromDomain(&center), err
}

func (c *Centers) getCenterByReferenceLegacy(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	reference := chi.URLParam(r, "reference")
	operator, err := c.operatorsService.GetCurrentOperator(r.Context())
	if err != nil {
		return nil, err
	}

	center, err := c.centersRepository.FindByOperatorAndUserReference(r.Context(), operator.UUID, reference)
	if err != nil {
		return nil, err
	}
	return center, err
}

func (c *Centers) getCenterByReference(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	reference := chi.URLParam(r, "reference")
	operator, err := c.operatorsService.GetCurrentOperator(r.Context())
	if err != nil {
		return nil, err
	}

	center, err := c.centersRepository.FindByOperatorAndUserReference(r.Context(), operator.UUID, reference)
	if err != nil {
		return nil, err
	}
	return model.CenterDTO{}.MapFromDomain(&center), err
}

// deleteCenterByReference deletes the center identified by the current operator and the reference.
func (c *Centers) deleteCenterByReference(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	reference := chi.URLParam(r, "reference")
	operator, err := c.operatorsService.GetCurrentOperator(r.Context())
	if err != nil {
		return nil, err
	}

	center, err := c.centersRepository.FindByOperatorAndUserReference(r.Context(), operator.UUID, reference)
	if err != nil {
		return nil, err
	}

	return nil, c.centersRepository.Delete(r.Context(), center)
}

// createBugReport creates a bug report for the specified center
func (c *Centers) createBugReport(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	uuid := chi.URLParam(r, "uuid")
	var request model.CreateBugReportRequestDTO
	if err := api.ParseRequestBody(r, c.validate, &request); err != nil {
		return nil, err
	}

	_, err := c.bugReportsService.CreateBugReport(r.Context(), uuid, request.Subject, request.Message)
	return nil, err
}

func (c *Centers) deleteCenterByUUID(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	centerUUID := chi.URLParam(r, "uuid")
	logrus.WithField("uuid", centerUUID).Trace("deleteCenterByUUID")

	operator, err := c.operatorsService.GetCurrentOperator(r.Context())
	if err != nil {
		return nil, err
	}

	center, err := c.centersRepository.FindByUUID(r.Context(), centerUUID)
	if err != nil {
		return nil, err
	}

	if center.OperatorUUID != operator.UUID && !security.HasRole(r.Context(), security.RoleAdmin) {
		return nil, security.ErrForbidden
	}

	return nil, c.centersRepository.Delete(r.Context(), center)
}

func (*Centers) getSearchParameters(r *http.Request) repositories.SearchParameters {
	result := repositories.SearchParameters{}
	appointmentParameter, hasAppointment := r.URL.Query()["appointment"]
	if hasAppointment {
		if tmp, ok := domain.ParseAppointmentType(appointmentParameter[0]); ok {
			result.Appointment = &tmp
		}
	}

	dccParameter, hasDcc := r.URL.Query()["dcc"]
	if hasDcc {
		if tmp, err := strconv.ParseBool(dccParameter[0]); err == nil {
			result.DCC = &tmp
		}
	}

	testKindParameter, hasTestKind := r.URL.Query()["kind"]
	if hasTestKind {
		if tmp, ok := domain.ParseTestKind(testKindParameter[0]); ok {
			result.TestKind = &tmp
		}
	}

	return result
}

func (*Centers) getBoundsParameter(r *http.Request) (domain.Bounds, bool, error) {
	var bounds domain.Bounds
	var ok bool
	var err error

	bounds.NorthEast.Latitude, ok, err = api.GetFloatParameter(r, "latne")
	if !ok || err != nil {
		return bounds, false, err
	}

	bounds.NorthEast.Longitude, ok, err = api.GetFloatParameter(r, "lngne")
	if !ok || err != nil {
		return bounds, false, err
	}

	bounds.SouthWest.Latitude, ok, err = api.GetFloatParameter(r, "latsw")
	if !ok || err != nil {
		return bounds, false, err
	}

	bounds.SouthWest.Longitude, ok, err = api.GetFloatParameter(r, "lngsw")
	if !ok || err != nil {
		return bounds, false, err
	}

	return bounds, true, nil
}
