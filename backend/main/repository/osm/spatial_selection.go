package osmdata

import (
	"context"
	"fmt"
	"math"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/paulmach/osm"
)

func bufferedBounds(b logistics.Bounds, snapKM float64) logistics.Bounds {

	latitude := snapKM / 111.19508
	south := math.Max(-90, b.South-latitude)
	north := math.Min(90, b.North+latitude)
	scale := math.Cos(math.Max(math.Abs(south), math.Abs(north)) * math.Pi / 180)
	longitude := 180.0

	if scale > 1e-9 {
		longitude = latitude / scale
	}

	return logistics.Bounds{West: math.Max(-180, b.West-longitude), South: south, East: math.Min(180, b.East+longitude), North: north}
}

func inside(b logistics.Bounds, c logistics.Coordinates) bool {
	return c.Longitude >= b.West && c.Longitude <= b.East && c.Latitude >= b.South && c.Latitude <= b.North
}

// Index only IDs inside the area. Keep complete ways later, including their
// outside nodes, so selection never connects nonadjacent nodes across a gap.
func selectAreaNodes(ctx context.Context, path string, bounds logistics.Bounds, limit int) (map[osm.NodeID]bool, error) {

	if err := bounds.Validate(); err != nil {
		return nil, fmt.Errorf("OSM selection bounds: %w", err)
	}

	nodes := map[osm.NodeID]bool{}
	err := scan(ctx, path, nodeObjects, func(object osm.Object) bool {
		node, ok := object.(*osm.Node)

		return ok && inside(bounds, logistics.Coordinates{Longitude: node.Lon, Latitude: node.Lat})
	}, func(object osm.Object) error {
		node, ok := object.(*osm.Node)

		if !ok || !inside(bounds, logistics.Coordinates{Longitude: node.Lon, Latitude: node.Lat}) {
			return nil
		}

		nodes[node.ID] = true

		if len(nodes) > limit {
			return fmt.Errorf("OSM area node selection exceeds maximum_objects=%d; select a smaller area", limit)
		}

		return nil
	})

	return nodes, err
}

func touchesArea(way *osm.Way, nodes map[osm.NodeID]bool) bool {

	for _, node := range way.Nodes {
		if nodes[node.ID] {
			return true
		}
	}

	return false
}
