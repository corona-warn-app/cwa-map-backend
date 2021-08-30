package model

import (
	"com.t-systems-mms.cwa/domain"
	"com.t-systems-mms.cwa/external/geocoding"
)

type GeocodeResultDTO struct {
	Address *string    `json:"address"`
	Bounds  *BoundsDTO `json:"bounds"`
}

func (g GeocodeResultDTO) MapFromModel(result *geocoding.Result) *GeocodeResultDTO {
	if result == nil {
		return nil
	}

	g.Address = &result.Address
	g.Bounds = BoundsDTO{}.MapFromModel(&result.Bounds)
	return &g
}

type BoundsDTO struct {
	NorthEast *CoordinatesDTO `json:"northEast"`
	SouthWest *CoordinatesDTO `json:"southWest"`
}

func (b BoundsDTO) MapFromModel(bounds *domain.Bounds) *BoundsDTO {
	if bounds == nil {
		return nil
	}

	b.NorthEast = CoordinatesDTO{}.MapFromModel(&bounds.NorthEast)
	b.SouthWest = CoordinatesDTO{}.MapFromModel(&bounds.SouthWest)
	return &b
}

type CoordinatesDTO struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func (c CoordinatesDTO) MapFromModel(coordinates *domain.Coordinates) *CoordinatesDTO {
	if coordinates == nil {
		return nil
	}

	c.Longitude = coordinates.Longitude
	c.Latitude = coordinates.Latitude
	return &c
}
