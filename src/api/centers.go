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
	"encoding/csv"
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"io"
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
}

func NewCentersAPI(centersService services.Centers, centersRepository repositories.Centers,
	operatorsService services.Operators, geocoder geocoding.Geocoder, auth *jwtauth.JWTAuth) *Centers {
	centers := &Centers{
		Router:            chi.NewRouter(),
		centersService:    centersService,
		centersRepository: centersRepository,
		operatorsService:  operatorsService,
		geocoder:          geocoder,
	}

	// public endpoints
	centers.Get("/", api.Handle(centers.FindCenters))
	centers.Get("/bounds", api.Handle(centers.Geocode))

	centers.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth))
		r.Use(jwtauth.Authenticator)

		r.Get("/all", api.Handle(centers.GetCenters))
		r.Post("/csv", api.Handle(centers.PrepareCsvImport))
		r.Post("/", api.Handle(centers.ImportCenters))

		// get centers
		r.Get("/reference/{reference}", api.Handle(centers.FindCenterByReference))

		// delete centers
		r.Delete("/{uuid}", api.Handle(centers.DeleteCenter))
		r.Delete("/reference/{reference}", api.Handle(centers.deleteCenterByReference))

		r.Route("/admin", func(r chi.Router) {
			r.Use(api.RequireRole(security.RoleAdmin))
			r.Get("/csv", centers.AdminGetCentersCSV)
			r.Post("/geocode", api.Handle(centers.AdminGeocodeAll))
		})
	})
	return centers
}

func (c *Centers) Geocode(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
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

func (c *Centers) FindCenters(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
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

func (c *Centers) PrepareCsvImport(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	parser := &services.CsvParser{}
	result, err := parser.Parse(r.Body)
	if parseError, isParseError := err.(*csv.ParseError); isParseError {
		return nil, api.HandlerError{
			Status: http.StatusBadRequest,
			Err:    parseError.Error(),
		}
	} else if err != nil {
		return nil, err
	}

	return model.MapToImportCenterResultDTOs(result), nil
}

func (c *Centers) GetCenters(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	operator, err := c.operatorsService.GetCurrentOperator(r.Context())
	if err != nil {
		return nil, err
	}

	centers, err := c.centersRepository.FindByOperator(r.Context(), operator.UUID, repositories.ParsePageRequest(r))
	if err != nil {
		return nil, err
	}
	return model.PageCenterDTO{
		PagedResult: api.PagedResult{Count: centers.Count},
		Result:      model.MapToCenterDTOs(centers.Result),
	}, nil
}

func (c *Centers) AdminGeocodeAll(_ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	centers, err := c.centersRepository.FindAll()
	if err != nil {
		return nil, err
	}
	go c.centersService.PerformGeocoding(centers)
	return nil, nil
}

func (c *Centers) AdminGetCentersCSV(w http.ResponseWriter, r *http.Request) {
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

	if err := csvWriter.Write([]string{"subject", "operator", "uuid", "name", "address", "zip", "region", "dcc", "testkinds", "appointment", "longitude", "latitude", "message"}); err != nil {
		logrus.WithError(err).Error("Error writing response")
		return
	}

	for _, center := range centers {
		if err := csvWriter.Write([]string{
			*center.Operator.Subject,
			center.Operator.Name,
			center.UUID,
			center.Name,
			center.Address,
			util.PtrToString(center.Zip, ""),
			util.PtrToString(center.Region, ""),
			util.BoolToString(center.DCC, "false"),
			strings.Join(center.TestKinds, ","),
			strconv.FormatFloat(center.Longitude, 'f', 10, 64),
			strconv.FormatFloat(center.Latitude, 'f', 10, 64),
			util.PtrToString((*string)(center.Appointment), ""),
			util.PtrToString(center.Message, ""),
		}); err != nil {
			logrus.WithError(err).Error("Error writing response")
			return
		}
	}
	csvWriter.Flush()
}

func (c *Centers) ImportCenters(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	var importData model.ImportCenterRequest
	buffer, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(buffer, &importData); err != nil {
		return nil, err
	}

	centers := make([]domain.Center, len(importData.Centers))
	for i, center := range importData.Centers {
		centers[i] = center.MapToDomain()
	}

	result, err := c.centersService.ImportCenters(r.Context(), centers, importData.DeleteAll)
	if err != nil {
		return nil, err
	}
	return model.MapToCenterDTOs(result), nil
}

func (c *Centers) FindCenterByReference(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
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

func (c *Centers) DeleteCenter(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	centerUUID := chi.URLParam(r, "uuid")
	logrus.WithField("uuid", centerUUID).Trace("DeleteCenter")

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

func (*Centers) getSearchDistance(r *http.Request, defaultDistance float64) float64 {
	if distanceParamer, hasDistance := r.URL.Query()["distance"]; hasDistance {
		if distance, err := strconv.ParseFloat(distanceParamer[0], 64); err == nil {
			return distance
		} else {
			logrus.WithFields(logrus.Fields{
				"parameter": "distance",
				"value":     distanceParamer[0],
			}).WithError(err).Warn("Invalid parameter format")
		}
	}
	return defaultDistance
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
