package graph_build_service

import (
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
)

func contains(b logistics.Bounds, c logistics.Coordinates) bool {
	return c.Longitude >= b.West && c.Longitude <= b.East && c.Latitude >= b.South && c.Latitude <= b.North
}

// condense merges spatial clusters of facilities into single weighted nodes
// until the graph fits the configured node budget. Clusters come from a grid
// whose cell grows until enough facilities share a cell; every merged node
// keeps the union of modes and external identities, its weight counts the
// merged facilities, and edges between members of one cluster are dropped.
func condense(g *logistics.LogisticsGraph, maximumNodes int) {

	if maximumNodes <= 0 || len(g.Nodes) <= maximumNodes || g.Bounds == nil {
		return
	}

	midLat := (g.Bounds.South + g.Bounds.North) / 2
	lonScale := math.Cos(midLat * math.Pi / 180)

	// Grow the cell until the clusters fit the budget, then shrink it while
	// they still fit, so the graph keeps as much detail as the budget allows.
	// The floor keeps coincident facilities from looping forever.
	cell := math.Max((g.Bounds.East-g.Bounds.West)*lonScale, g.Bounds.North-g.Bounds.South) / math.Sqrt(float64(maximumNodes))

	for clusterCount(g.Nodes, cell, lonScale) > maximumNodes {
		cell *= 1.2
	}

	for cell > 0.0005 {
		finer := cell / 1.2

		if clusterCount(g.Nodes, finer, lonScale) > maximumNodes {
			break
		}

		cell = finer
	}

	type cellKey struct{ lat, lon int }

	members := map[cellKey][]int{}
	clusterOf := map[string]string{}

	for i, node := range g.Nodes {
		key := cellKey{int(math.Floor(node.Coordinates.Latitude / cell)), int(math.Floor(node.Coordinates.Longitude * lonScale / cell))}
		members[key] = append(members[key], i)
		clusterOf[node.ID] = fmt.Sprintf("cluster:%d:%d", key.lat, key.lon)
	}

	condensed := make([]logistics.LogisticsNode, 0, len(members))

	for _, indexes := range members {
		node := logistics.LogisticsNode{ID: clusterOf[g.Nodes[indexes[0]].ID], RegionID: g.Nodes[indexes[0]].RegionID, Coordinates: centroid(g.Nodes, indexes), Weight: len(indexes)}

		kinds := map[string]int{}

		for _, i := range indexes {
			member := g.Nodes[i]
			node.Modes = append(node.Modes, member.Modes...)
			node.ExternalIDs = append(node.ExternalIDs, member.ExternalIDs...)
			node.Sources = append(node.Sources, member.Sources...)
			kinds[member.Kind]++

			if node.Name == "" {
				node.Name = member.Name
			}

			if node.Cargo == "" {
				node.Cargo = member.Cargo
			}
		}

		for _, kind := range slices.Sorted(maps.Keys(kinds)) {
			if kinds[kind] > kinds[node.Kind] {
				node.Kind = kind
			}
		}

		slices.Sort(node.Modes)
		node.Modes = slices.Compact(node.Modes)
		slices.SortFunc(node.ExternalIDs, func(a, b logistics.ExternalID) int {
			return cmpText(string(a.Namespace)+":"+a.Value, string(b.Namespace)+":"+b.Value)
		})
		node.ExternalIDs = slices.Compact(node.ExternalIDs)
		slices.Sort(node.Sources)
		node.Sources = slices.Compact(node.Sources)

		if node.Name == "" {
			node.Name = node.ID
		}

		if node.Kind == "" {
			node.Kind = "freight_facility"
		}

		condensed = append(condensed, node)
	}

	slices.SortFunc(condensed, func(a, b logistics.LogisticsNode) int {
		return cmpText(a.ID, b.ID)
	})

	type edgeKey struct {
		from, to string
		mode     logistics.TransportMode
	}

	edges := map[edgeKey]logistics.LogisticsEdge{}

	for _, edge := range g.Edges {
		from, to := clusterOf[edge.FromNodeID], clusterOf[edge.ToNodeID]

		if from == "" || to == "" || from == to {
			continue
		}

		key := edgeKey{from, to, edge.Mode}

		if existing, ok := edges[key]; ok && existing.DistanceKM <= edge.DistanceKM {
			continue
		}

		edge.ID = fmt.Sprintf("%s:%s:%s", edge.Mode, from, to)
		edge.FromNodeID = from
		edge.ToNodeID = to
		edges[key] = edge
	}

	g.Coverage.Warnings = append(g.Coverage.Warnings, fmt.Sprintf("Graph condensed: %d facilities merged into %d nodes", len(g.Nodes), len(condensed)))
	g.Nodes = condensed

	orderedEdges := make([]logistics.LogisticsEdge, 0, len(edges))

	for _, edge := range edges {
		orderedEdges = append(orderedEdges, edge)
	}

	slices.SortFunc(orderedEdges, func(a, b logistics.LogisticsEdge) int {
		return cmpText(a.ID, b.ID)
	})

	g.Edges = orderedEdges
}

func clusterCount(nodes []logistics.LogisticsNode, cell, lonScale float64) int {

	seen := map[[2]int]bool{}

	for _, node := range nodes {
		seen[[2]int{int(math.Floor(node.Coordinates.Latitude / cell)), int(math.Floor(node.Coordinates.Longitude * lonScale / cell))}] = true
	}

	return len(seen)
}

func centroid(nodes []logistics.LogisticsNode, indexes []int) logistics.Coordinates {

	var c logistics.Coordinates

	for _, i := range indexes {
		c.Longitude += nodes[i].Coordinates.Longitude
		c.Latitude += nodes[i].Coordinates.Latitude
	}

	c.Longitude /= float64(len(indexes))
	c.Latitude /= float64(len(indexes))

	return c
}

func cmpText(a, b string) int {
	if a < b {
		return -1
	}

	if a > b {
		return 1
	}

	return 0
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
