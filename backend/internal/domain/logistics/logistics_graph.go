package logistics

import (
	"fmt"
	"slices"
)

type LogisticsGraph struct {
	Coverage Coverage        `json:"coverage"`
	ID       string          `json:"id"`
	RegionID string          `json:"region_id"`
	Nodes    []LogisticsNode `json:"nodes"`
	Edges    []LogisticsEdge `json:"edges"`
}

func (g LogisticsGraph) Validate() error {
	if err := validateText("graph ID", g.ID); err != nil {
		return err
	}
	if err := validateText("graph region ID", g.RegionID); err != nil {
		return err
	}
	nodes := make(map[string]bool)
	nodeModes := make(map[string][]TransportMode)
	for _, node := range g.Nodes {
		if err := node.Validate(); err != nil {
			return err
		}
		if node.RegionID != g.RegionID || nodes[node.ID] {
			return fmt.Errorf("invalid or duplicate graph node %q", node.ID)
		}
		nodes[node.ID] = true
		nodeModes[node.ID] = node.Modes
	}
	edges := make(map[string]bool)
	for _, edge := range g.Edges {
		if err := validateText("edge ID", edge.ID); err != nil {
			return err
		}
		if edges[edge.ID] || !nodes[edge.FromNodeID] || !nodes[edge.ToNodeID] {
			return fmt.Errorf("invalid edge references %q", edge.ID)
		}
		if edge.FromNodeID == edge.ToNodeID || !slices.Contains(nodeModes[edge.FromNodeID], edge.Mode) || !slices.Contains(nodeModes[edge.ToNodeID], edge.Mode) || !edge.Mode.Valid() || !inRange(edge.DistanceKM, 0, 1e9) || len(edge.Geometry) < 2 {
			return fmt.Errorf("invalid edge attributes %q", edge.ID)
		}
		for _, point := range edge.Geometry {
			if err := point.Validate(); err != nil {
				return err
			}
		}
		edges[edge.ID] = true
	}
	return nil
}
