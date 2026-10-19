package graph_build_service

import (
	"context"
	"crypto/rand"
	"fmt"
	"sort"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
)

type Service struct {
	Source Source
	Policy logistics.FreightPolicy
}

func (p Service) Build(ctx context.Context, region logistics.Region) (logistics.LogisticsGraph, error) {
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
	g := logistics.LogisticsGraph{Bounds: &region.Bounds, ID: rand.Text(), RegionID: region.ID, Nodes: []logistics.LogisticsNode{}, Edges: []logistics.LogisticsEdge{}}
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
