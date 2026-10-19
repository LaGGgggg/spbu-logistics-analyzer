package logistics

import (
	"fmt"
	"math"
	"strings"
)

func validateText(field, value string) error {

	if value == "" || strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must be nonempty and have no surrounding whitespace", field)
	}

	return nil
}

func inRange(value, min, max float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= min && value <= max
}
