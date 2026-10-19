package graph_build_service

import (
	"slices"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
)

func contains(b logistics.Bounds, c logistics.Coordinates) bool {
	return c.Longitude >= b.West && c.Longitude <= b.East && c.Latitude >= b.South && c.Latitude <= b.North
}

func mergeNode(g *logistics.LogisticsGraph, index map[string]int, node logistics.LogisticsNode) {

	keys := []string{"id:" + node.ID}

	for _, id := range node.ExternalIDs {
		value := id.Value

		if id.Namespace == logistics.ExternalESR {
			value = logistics.ESRKey(value)

			if value == "" {
				continue
			}
		}

		keys = append(keys, "external:"+string(id.Namespace)+":"+value)
	}

	found := -1

	for _, key := range keys {
		if i, ok := index[key]; ok {
			found = i
			break
		}
	}

	if found < 0 {
		found = len(g.Nodes)
		g.Nodes = append(g.Nodes, node)
	} else {
		n := &g.Nodes[found]
		n.Modes = append(n.Modes, node.Modes...)
		slices.Sort(n.Modes)
		n.Modes = slices.Compact(n.Modes)

		for _, id := range node.ExternalIDs {
			if !slices.Contains(n.ExternalIDs, id) {
				n.ExternalIDs = append(n.ExternalIDs, id)
			}
		}

		for _, source := range node.Sources {
			if !slices.Contains(n.Sources, source) {
				n.Sources = append(n.Sources, source)
			}
		}
	}

	for _, key := range keys {
		index[key] = found
	}
}
