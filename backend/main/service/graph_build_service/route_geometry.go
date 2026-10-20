package graph_build_service

import (
	"context"
	"math"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
)

// Keep short paths intact. Long paths use iterative Douglas-Peucker in a local
// metre projection. Distances still come from the original routing network.
func simplifyRoute(ctx context.Context, points []logistics.Coordinates, toleranceM float64) ([]logistics.Coordinates, error) {

	if toleranceM == 0 || len(points) <= 32 {
		return points, nil
	}

	minLat, maxLat := points[0].Latitude, points[0].Latitude

	for _, point := range points {
		minLat = math.Min(minLat, point.Latitude)
		maxLat = math.Max(maxLat, point.Latitude)
	}

	latitude := math.Min(math.Abs(minLat), math.Abs(maxLat))

	if minLat <= 0 && maxLat >= 0 {
		latitude = 0
	}

	scaleX := 111195.08 * math.Cos(latitude*math.Pi/180)
	const scaleY = 111195.08
	keep := make([]bool, len(points))
	keep[0], keep[len(points)-1] = true, true
	type segment struct{ first, last int }
	stack := []segment{{0, len(points) - 1}}
	threshold := toleranceM * toleranceM

	for len(stack) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		a, b := points[current.first], points[current.last]
		dx := (b.Longitude - a.Longitude) * scaleX
		dy := (b.Latitude - a.Latitude) * scaleY
		lengthSquared := dx*dx + dy*dy
		farthest := -1
		maximum := threshold

		for i := current.first + 1; i < current.last; i++ {
			x := (points[i].Longitude - a.Longitude) * scaleX
			y := (points[i].Latitude - a.Latitude) * scaleY
			projection := 0.0

			if lengthSquared > 0 {
				projection = math.Max(0, math.Min(1, (x*dx+y*dy)/lengthSquared))
			}

			x -= projection * dx
			y -= projection * dy
			distanceSquared := x*x + y*y

			if distanceSquared > maximum {
				maximum = distanceSquared
				farthest = i
			}
		}

		if farthest >= 0 {
			keep[farthest] = true
			stack = append(stack, segment{current.first, farthest}, segment{farthest, current.last})
		}
	}

	result := make([]logistics.Coordinates, 0)

	for i, point := range points {
		if keep[i] {
			result = append(result, point)
		}
	}

	return result, nil
}
