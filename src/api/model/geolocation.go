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
