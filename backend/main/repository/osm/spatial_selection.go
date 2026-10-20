package osmdata

import (
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/graph_build_service"
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

type nodeIndex []osm.NodeID

func (index nodeIndex) contains(id osm.NodeID) bool {
	_, found := slices.BinarySearch(index, id)

	return found
}

// Index only IDs inside the area. Keep complete ways later, including their
// outside nodes, so selection never connects nonadjacent nodes across a gap.
func selectAreaNodes(ctx context.Context, path string, bounds logistics.Bounds, limit int) (nodeIndex, error) {

	if err := bounds.Validate(); err != nil {
		return nil, fmt.Errorf("OSM selection bounds: %w", err)
	}

	nodes := nodeIndex{}
	err := scan(ctx, path, nodeObjects, func(object osm.Object) bool {
		node, ok := object.(*osm.Node)

		return ok && inside(bounds, logistics.Coordinates{Longitude: node.Lon, Latitude: node.Lat})
	}, func(object osm.Object) error {
		node, ok := object.(*osm.Node)

		if !ok || !inside(bounds, logistics.Coordinates{Longitude: node.Lon, Latitude: node.Lat}) {
			return nil
		}

		nodes = append(nodes, node.ID)

		if len(nodes) > limit {
			return fmt.Errorf("%w: OSM area index exceeds maximum_index_nodes=%d", graph_build_service.ErrDatasetLimit, limit)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	slices.Sort(nodes)

	return slices.Compact(nodes), nil
}

func touchesArea(way *osm.Way, nodes nodeIndex) bool {

	for _, node := range way.Nodes {
		if nodes.contains(node.ID) {
			return true
		}
	}

	return false
}
