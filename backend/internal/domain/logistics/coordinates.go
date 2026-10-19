package logistics

import "fmt"

type Coordinates struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func (c Coordinates) Validate() error {

	if !inRange(c.Longitude, -180, 180) || !inRange(c.Latitude, -90, 90) {
		return fmt.Errorf("invalid WGS84 coordinates: longitude=%v latitude=%v", c.Longitude, c.Latitude)
	}

	return nil
}
