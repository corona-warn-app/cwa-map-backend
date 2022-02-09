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

package geocoding

import (
	"com.t-systems-mms.cwa/domain"
	"context"
	"errors"
)

type Result struct {
	Address     string
	Bounds      domain.Bounds
	Coordinates domain.Coordinates
	Zip         string
	Region      string
}

var (
	ErrNoResult       = errors.New("no results")
	ErrTooManyResults = errors.New("too many results")
)

// Geocoder interface provides common functions for different geocoding implementations.
type Geocoder interface {
	// GetCoordinates resolves the given address to geographical coordinates.
	//
	// address could be postal codes, cities or complete addresses
	GetCoordinates(ctx context.Context, address string) (Result, error)
}
