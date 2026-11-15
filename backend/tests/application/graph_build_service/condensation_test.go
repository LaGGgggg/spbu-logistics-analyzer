package graph_build_service_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/graph_build_service"
)

func TestGraphCondensationMergesClusters(t *testing.T) {

	reference := func(id string, lon, lat float64, mode logistics.TransportMode) logistics.Reference {
		c := logistics.Coordinates{Longitude: lon, Latitude: lat}
		return logistics.Reference{ExternalIDs: []logistics.ExternalID{{Namespace: logistics.ExternalOSMNode, Value: id}}, Modes: []logistics.TransportMode{mode}, Cargo: true, Source: "test", Coordinates: &c}
	}

	point := func(id string, lon, lat float64) logistics.TrackPoint {
		return logistics.TrackPoint{ID: id, Coordinates: logistics.Coordinates{Longitude: lon, Latitude: lat}}
	}

	track := logistics.Track{ID: "osm_way:link", Mode: logistics.TransportRail, Points: []logistics.TrackPoint{
		point("p1", 30.000, 59.5),
		point("p2", 30.25, 60.2),
		point("p3", 30.5, 60.9),
	}}

	dataset := logistics.Dataset{
		Tracks: []logistics.Track{track},
		References: []logistics.Reference{
			reference("near-a", 30.000, 59.5, logistics.TransportRail),
			reference("near-b", 30.0001, 59.5, logistics.TransportSea),
			reference("far-c", 30.5, 60.9, logistics.TransportRail),
		},
	}

	policy := logistics.FreightPolicy{MaximumSnapKM: 1, Optimization: 0, MaximumNodes: 2}
	p := graph_build_service.Service{Policy: policy, Source: fixtureSource{dataset}}
	g, err := p.Build(context.Background(), logistics.Region{ID: "r", Name: "r", Bounds: logistics.Bounds{West: 29, East: 31, South: 59, North: 61}})

	if err != nil {
		t.Fatal(err)
	}

	if len(g.Nodes) != 2 {
		t.Fatalf("nodes=%+v", g.Nodes)
	}

	totalWeight := 0

	for _, node := range g.Nodes {
		if !strings.HasPrefix(node.ID, "cluster:") {
			t.Fatalf("uncondensed node %s", node.ID)
		}

		totalWeight += node.Weight
	}

	if totalWeight != 3 {
		t.Fatalf("total weight=%d", totalWeight)
	}

	merged := g.Nodes[slices.IndexFunc(g.Nodes, func(n logistics.LogisticsNode) bool { return n.Weight == 2 })]

	if merged.ID == "" || !slices.Contains(merged.Modes, logistics.TransportSea) || !slices.Contains(merged.Modes, logistics.TransportRail) {
		t.Fatalf("merged node modes=%+v", merged)
	}

	badEdge := slices.ContainsFunc(g.Edges, func(e logistics.LogisticsEdge) bool {
		return e.Mode != logistics.TransportRail || !strings.HasPrefix(e.FromNodeID, "cluster:") || !strings.HasPrefix(e.ToNodeID, "cluster:") || e.FromNodeID == e.ToNodeID
	})

	if len(g.Edges) != 2 || badEdge {
		t.Fatalf("edges=%+v", g.Edges)
	}

	if !slices.ContainsFunc(g.Coverage.Warnings, func(w string) bool { return strings.Contains(w, "condensed") }) {
		t.Fatalf("warnings=%+v", g.Coverage.Warnings)
	}

	if g.Metrics == nil || len(g.Metrics.PageRank) != 2 {
		t.Fatalf("metrics=%+v", g.Metrics)
	}
}

func TestGraphCondensationDisabled(t *testing.T) {

	c := logistics.Coordinates{Longitude: 30, Latitude: 60}
	refs := []logistics.Reference{{ExternalIDs: []logistics.ExternalID{{Namespace: logistics.ExternalOSMNode, Value: "1"}}, Modes: []logistics.TransportMode{logistics.TransportRail}, Cargo: true, Source: "test", Coordinates: &c}}
	p := graph_build_service.Service{Policy: logistics.FreightPolicy{MaximumSnapKM: 1, Optimization: 0}, Source: fixtureSource{logistics.Dataset{References: refs}}}
	g, err := p.Build(context.Background(), logistics.Region{ID: "r", Name: "r", Bounds: logistics.Bounds{West: 29, East: 31, South: 59, North: 61}})

	if err != nil {
		t.Fatal(err)
	}

	if len(g.Nodes) != 1 || g.Nodes[0].ID != "osm_node:1" || g.Nodes[0].Weight != 1 {
		t.Fatalf("nodes=%+v", g.Nodes)
	}
}
