package services

import (
	"com.t-systems-mms.cwa/core"
	"com.t-systems-mms.cwa/core/security"
	"com.t-systems-mms.cwa/core/util"
	"com.t-systems-mms.cwa/domain"
	"com.t-systems-mms.cwa/external/geocoding"
	"com.t-systems-mms.cwa/repositories"
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"math/rand"
)

type ImportCenterResult struct {
	Center   domain.Center
	Warnings []string
	Errors   []string
}

var (
	ErrDuplicateUserReference = core.ApplicationError("duplicate user reference")
)

type Centers interface {
	ImportCenters(ctx context.Context, centers []domain.Center, deleteAll bool) ([]domain.Center, error)
	Save(ctx context.Context, center *domain.Center, geocoding bool) error
	PerformGeocoding(centers []domain.Center)
}

type centersService struct {
	centersRepository repositories.Centers
	operators         repositories.Operators
	operatorsService  Operators
	geocoder          geocoding.Geocoder
	validate          *validator.Validate
}

func NewCentersService(centersRepository repositories.Centers, operators repositories.Operators, operatorsService Operators, geocoder geocoding.Geocoder) Centers {
	validate := validator.New()
	validate.RegisterTagNameFunc(util.JsonTagNameFunc)

	return &centersService{
		centersRepository: centersRepository,
		operators:         operators,
		operatorsService:  operatorsService,
		geocoder:          geocoder,
		validate:          validate,
	}
}

func (s *centersService) Save(ctx context.Context, center *domain.Center, geocoding bool) error {
	operator, err := s.operatorsService.GetCurrentOperator(ctx)
	if err != nil {
		return err
	}

	if err := s.validate.Struct(center); err != nil {
		return err
	}

	if !security.HasRole(ctx, security.RoleDCC) {
		tmp := false
		center.DCC = &tmp
	}

	if util.IsNotNilOrEmpty(center.UserReference) {
		if existing, err := s.centersRepository.FindByOperatorAndUserReference(ctx, operator.UUID, *center.UserReference); err == nil {
			if util.IsNotNilOrEmpty(&center.UUID) && existing.UUID != center.UUID {
				return ErrDuplicateUserReference
			}

			// If there is already a center with this userReference, use its UUID to replace it
			center.UUID = existing.UUID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	center.OperatorUUID = operator.UUID
	center.Ranking = rand.Float64()
	if err := s.centersRepository.Save(ctx, center); err == nil {
		if geocoding {
			return s.GeocodeCenter(ctx, center)
		}
		return nil
	} else {
		return err
	}
}

func (s *centersService) ImportCenters(ctx context.Context, centers []domain.Center, deleteAll bool) ([]domain.Center, error) {
	// validate each center before
	for _, center := range centers {
		if err := s.validate.Struct(center); err != nil {
			return nil, err
		}
	}

	operator, err := s.operatorsService.GetCurrentOperator(ctx)
	if err != nil {
		return nil, err
	}

	err = s.centersRepository.UseTransaction(ctx, func(ctx context.Context) error {
		if deleteAll {
			if err := s.centersRepository.DeleteByOperator(ctx, operator.UUID); err != nil {
				return err
			}
		}

		for i, _ := range centers {
			if err := s.Save(ctx, &centers[i], false); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	go s.PerformGeocoding(centers)
	return centers, err
}

func (s *centersService) GeocodeCenter(ctx context.Context, center *domain.Center) error {
	logrus.WithFields(logrus.Fields{
		"center":  center.UUID,
		"address": center.Address,
	}).Trace("Geocoding center")

	g, err := s.geocoder.GetCoordinates(ctx, center.Address)
	if err != nil {
		logrus.
			WithFields(logrus.Fields{
				"center":  center.UUID,
				"address": center.Address,
			}).
			WithError(err).
			Error("Error geocoding center")

		if err == geocoding.ErrTooManyResults || err == geocoding.ErrNoResult {
			msg := fmt.Sprintf("Geocoding: %s", err.Error())
			center.Message = &msg
		}
	} else {
		center.Zip = &g.Zip
		center.Region = &g.Region
		center.Coordinates = domain.Coordinates{
			Longitude: g.Coordinates.Longitude,
			Latitude:  g.Coordinates.Latitude,
		}
	}

	err = s.centersRepository.Save(context.Background(), center)
	if err != nil {
		logrus.WithError(err).Error("Error saving center")
	}
	return err
}

func (s *centersService) PerformGeocoding(centers []domain.Center) {
	logrus.WithFields(logrus.Fields{
		"count": len(centers),
	}).Info("Starting geocoding of importet centers")

	for _, center := range centers {
		uuid := center.UUID
		center, err := s.centersRepository.FindByUUID(context.Background(), center.UUID)
		if err != nil {
			logrus.
				WithFields(logrus.Fields{"center": uuid}).
				WithError(err).
				Warn("Center to found, maybe already deleted")
			continue
		}

		g, err := s.geocoder.GetCoordinates(context.Background(), center.Address)
		if err != nil {
			logrus.
				WithFields(logrus.Fields{
					"center":  center.UUID,
					"address": center.Address,
				}).
				WithError(err).
				Error("Error geocoding center")

			if err == geocoding.ErrTooManyResults || err == geocoding.ErrNoResult {
				msg := fmt.Sprintf("Geocoding: %s", err.Error())
				center.Message = &msg
			}
		} else {
			center.Zip = &g.Zip
			center.Region = &g.Region
			center.Coordinates = domain.Coordinates{
				Longitude: g.Coordinates.Longitude,
				Latitude:  g.Coordinates.Latitude,
			}
		}

		err = s.centersRepository.Save(context.Background(), &center)
		if err != nil {
			logrus.WithError(err).Error("Error saving center")
		}
	}
	logrus.WithFields(logrus.Fields{
		"count": len(centers),
	}).Info("Geocoding for centers completed")
}
