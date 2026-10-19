package network

import (
	"context"
	"crypto/rand"
	"fmt"
	"slices"
	"sort"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
)

type Builder interface {
	Build(context.Context, logistics.Region) (logistics.LogisticsGraph, error)
}

type Source interface {
	Load(context.Context, logistics.Region) (logistics.Dataset, error)
}
type Pipeline struct {
	Source Source
	Policy logistics.FreightPolicy
}

func (p Pipeline) Build(ctx context.Context, region logistics.Region) (logistics.LogisticsGraph, error) {
	if err := region.Validate(); err != nil {
		return logistics.LogisticsGraph{}, err
	}
	if err := p.Policy.Validate(); err != nil {
		return logistics.LogisticsGraph{}, err
	}
	if p.Source == nil {
		return logistics.LogisticsGraph{}, fmt.Errorf("no dataset source configured")
	}
	dataset, err := p.Source.Load(ctx, region)
	if err != nil {
		return logistics.LogisticsGraph{}, err
	}
	g := logistics.LogisticsGraph{ID: rand.Text(), RegionID: region.ID, Nodes: []logistics.LogisticsNode{}, Edges: []logistics.LogisticsEdge{}}
	g.Coverage = logistics.Coverage{Sources: dataset.Sources, Warnings: dataset.Warnings, UnmatchedReferenceIDs: []string{}, UnroutedNodeModes: []string{}}
	byID := map[string]int{}
	for _, f := range dataset.Features {
		if err := ctx.Err(); err != nil {
			return g, err
		}
		if !contains(region.Bounds, f.Coordinates) {
			continue
		}
		g.Coverage.Candidates++
		node, ok := p.Policy.Classify(f, dataset.References)
		if !ok {
			g.Coverage.Rejected++
			continue
		}
		if err := f.ExternalID.Validate(); err != nil {
			return g, err
		}
		node.RegionID = region.ID
		mergeNode(&g, byID, node)
	}
	observedNodes := append([]logistics.LogisticsNode(nil), g.Nodes...)
	for _, ref := range dataset.References {
		if !ref.Cargo {
			continue
		}
		if ref.Coordinates != nil && !contains(region.Bounds, *ref.Coordinates) {
			continue
		}
		g.Coverage.ReferenceTotal++
		matched := false
		for _, node := range observedNodes {
			for _, a := range node.ExternalIDs {
				for _, b := range ref.ExternalIDs {
					if logistics.SameExternalID(a, b) {
						matched = true
					}
				}
			}
		}
		if matched {
			g.Coverage.ReferenceMatched++
			continue
		}
		if len(ref.ExternalIDs) == 0 {
			return g, fmt.Errorf("reference without identity")
		}
		id := string(ref.ExternalIDs[0].Namespace) + ":" + ref.ExternalIDs[0].Value
		g.Coverage.UnmatchedReferenceIDs = append(g.Coverage.UnmatchedReferenceIDs, id)
		if ref.Coordinates == nil {
			continue
		}
		if ref.PhysicalID != "" {
			id = "physical:" + ref.PhysicalID
		}
		node := logistics.LogisticsNode{ID: id, RegionID: region.ID, Name: ref.Name, Coordinates: *ref.Coordinates, Modes: ref.Modes, ExternalIDs: ref.ExternalIDs, Sources: []string{ref.Source}, Kind: "freight_facility"}
		if node.Name == "" {
			node.Name = id
		}
		mergeNode(&g, byID, node)
	}
	sort.Slice(g.Nodes, func(i, j int) bool { return g.Nodes[i].ID < g.Nodes[j].ID })
	g.Coverage.Accepted = len(g.Nodes)
	if len(g.Nodes) == 0 {
		return g, fmt.Errorf("no confirmed freight facilities in region %s", region.ID)
	}
	if err := connect(ctx, &g, dataset.Tracks, p.Policy.MaximumSnapKM); err != nil {
		return g, err
	}
	if err := g.Validate(); err != nil {
		return g, err
	}
	return g, nil
}
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
