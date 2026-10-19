package logistics

import "fmt"

type Bounds struct {
	West  float64 `json:"west"`
	South float64 `json:"south"`
	East  float64 `json:"east"`
	North float64 `json:"north"`
}

func (b Bounds) Validate() error {

	if err := (Coordinates{Longitude: b.West, Latitude: b.South}).Validate(); err != nil {
		return fmt.Errorf("southwest bound: %w", err)
	}

	if err := (Coordinates{Longitude: b.East, Latitude: b.North}).Validate(); err != nil {
		return fmt.Errorf("northeast bound: %w", err)
	}

	if b.West >= b.East || b.South >= b.North {
		return fmt.Errorf("bounds must have west < east and south < north")
	}

	return nil
}
