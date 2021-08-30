package domain

type Bounds struct {
	NorthEast Coordinates
	SouthWest Coordinates
}

type Coordinates struct {
	Longitude float64
	Latitude  float64
}
