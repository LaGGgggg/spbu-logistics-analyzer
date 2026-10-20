package graph_build_service

import (
	"context"
	"math"
	"slices"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
)

type spatialPoint struct {
	id  string
	xyz [3]float64
}

type spatialIndex struct {
	point       spatialPoint
	axis        int
	left, right *spatialIndex
}

func xyz(c logistics.Coordinates) [3]float64 {

	lat, lon := c.Latitude*math.Pi/180, c.Longitude*math.Pi/180

	return [3]float64{math.Cos(lat) * math.Cos(lon), math.Cos(lat) * math.Sin(lon), math.Sin(lat)}
}

func newSpatialIndex(ctx context.Context, points []spatialPoint, depth int) *spatialIndex {

	if len(points) == 0 || ctx.Err() != nil {
		return nil
	}

	axis := depth % 3
	slices.SortFunc(points, func(a, b spatialPoint) int {
		if a.xyz[axis] < b.xyz[axis] {
			return -1
		}

		if a.xyz[axis] > b.xyz[axis] {
			return 1
		}

		if a.id < b.id {
			return -1
		}

		if a.id > b.id {
			return 1
		}

		return 0
	})
	mid := len(points) / 2

	return &spatialIndex{point: points[mid], axis: axis, left: newSpatialIndex(ctx, points[:mid], depth+1), right: newSpatialIndex(ctx, points[mid+1:], depth+1)}
}

func (index *spatialIndex) nearest(c logistics.Coordinates, maxKM float64) string {

	target := xyz(c)
	best := ""
	chord := 2 * math.Sin(maxKM/(2*6371.0088))
	bestSquared := chord * chord
	var visit func(*spatialIndex)
	visit = func(n *spatialIndex) {
		if n == nil {
			return
		}

		squared := 0.0

		for axis := range 3 {
			delta := n.point.xyz[axis] - target[axis]
			squared += delta * delta
		}

		if squared < bestSquared || (squared == bestSquared && (best == "" || n.point.id < best)) {
			bestSquared = squared
			best = n.point.id
		}

		delta := target[n.axis] - n.point.xyz[n.axis]
		near, far := n.left, n.right

		if delta > 0 {
			near, far = far, near
		}

		visit(near)

		if delta*delta <= bestSquared {
			visit(far)
		}
	}
	visit(index)

	return best
}
