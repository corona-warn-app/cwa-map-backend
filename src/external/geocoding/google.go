package geocoding

import (
	"com.t-systems-mms.cwa/core/util"
	"com.t-systems-mms.cwa/domain"
	"context"
	"github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
)

type GoogleGeocoder struct {
	config GoogleGeocoderConfig
	client *maps.Client
}

type GoogleGeocoderConfig struct {
	ApiKey  string
	Country string
}

func NewGoogleGeocoder(config GoogleGeocoderConfig) (*GoogleGeocoder, error) {
	client, err := maps.NewClient(maps.WithAPIKey(config.ApiKey))
	if err != nil {
		return nil, err
	}

	return &GoogleGeocoder{
		config: config,
		client: client,
	}, nil
}

func (g *GoogleGeocoder) GetCoordinates(ctx context.Context, address string) (Result, error) {
	logrus.WithFields(logrus.Fields{
		"address": address,
	}).Debug("GetCoordinates")

	results, err := g.client.Geocode(ctx, &maps.GeocodingRequest{
		Address: address,
		Region:  "de", //TODO move to configuration
		Components: map[maps.Component]string{
			"country": "de", //TODO move to configuration
		},
	})

	if err != nil {
		return Result{}, err
	}

	if len(results) == 0 {
		return Result{}, ErrNoResult
	}

	var result *maps.GeocodingResult
	if len(results) > 1 {
		for _, r := range results {
			if r.Types != nil && len(r.Types) > 0 && util.ArrayContainsOne(r.Types, "street_address") {
				result = &r
				break
			}
		}

		if result == nil {
			return Result{}, ErrTooManyResults
		}
	} else {
		result = &results[0]
	}

	return Result{
		Address: result.FormattedAddress,
		Region:  g.getAddressComponent(result, "administrative_area_level_1"),
		Zip:     g.getAddressComponent(result, "postal_code"),
		Bounds: domain.Bounds{
			NorthEast: domain.Coordinates{
				Longitude: result.Geometry.Viewport.NorthEast.Lng,
				Latitude:  result.Geometry.Viewport.NorthEast.Lat,
			},
			SouthWest: domain.Coordinates{
				Longitude: result.Geometry.Viewport.SouthWest.Lng,
				Latitude:  result.Geometry.Viewport.SouthWest.Lat,
			},
		},
		Coordinates: domain.Coordinates{
			Longitude: result.Geometry.Location.Lng,
			Latitude:  result.Geometry.Location.Lat,
		},
	}, nil
}

func (g *GoogleGeocoder) getAddressComponent(result *maps.GeocodingResult, component string) string {
	for _, c := range result.AddressComponents {
		if util.ArrayContainsOne(c.Types, component) {
			return c.LongName
		}
	}
	return ""
}
