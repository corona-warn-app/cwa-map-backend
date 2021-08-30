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
