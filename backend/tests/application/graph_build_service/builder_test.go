package graph_build_service_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/graph_build_service"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/repository/osm"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/tests/testutil"
)

func TestOSMPipelineFollowsInfrastructure(t *testing.T) {

	root := testutil.FixtureDir(t)
	pipeline := graph_build_service.Service{Policy: logistics.DefaultFreightPolicy(), Source: osmdata.Source{Catalog: osmdata.Catalog{MaximumObjects: 1000, Regions: map[string]osmdata.Files{"test": {OSM: filepath.Join(root, "freight.osm"), References: filepath.Join(root, "references.json")}}}}}
	region := logistics.Region{ID: "test", Name: "Test region", Bounds: logistics.Bounds{West: 29, South: 59, East: 31, North: 61}}
	g, err := pipeline.Build(context.Background(), region)

	if err != nil {
		t.Fatal(err)
	}

	if len(g.Nodes) != 3 {
		t.Fatalf("nodes=%+v", g.Nodes)
	}

	if len(g.Edges) != 1 {
		t.Fatalf("edges=%+v", g.Edges)
	}

	edge := g.Edges[0]

	if edge.FromNodeID != "osm_node:1" || edge.ToNodeID != "osm_node:4" || !edge.Directed {
		t.Fatalf("wrong direction: %+v", edge)
	}

	if len(edge.Geometry) != 4 || edge.DistanceKM < 2.7 {
		t.Fatalf("route shortcut: %+v", edge)
	}

	if g.Coverage.ReferenceMatched != 1 || g.Coverage.Rejected != 1 {
		t.Fatalf("coverage=%+v", g.Coverage)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := pipeline.Build(ctx, region); err == nil {
		t.Fatal("cancellation ignored")
	}
}

type fixtureSource struct{ dataset logistics.Dataset }

func (f fixtureSource) Load(context.Context, logistics.Region) (logistics.Dataset, error) {
	return f.dataset, nil
}

func TestPhysicalIdentityAndDisconnectedNodes(t *testing.T) {

	coord := logistics.Coordinates{Longitude: 30, Latitude: 60}
	refs := []logistics.Reference{{PhysicalID: "port", ExternalIDs: []logistics.ExternalID{{Namespace: logistics.ExternalOSMNode, Value: "1"}, {Namespace: logistics.ExternalOSMWay, Value: "2"}}, Modes: []logistics.TransportMode{logistics.TransportRoad, logistics.TransportSea}, Cargo: true, Source: "test"}}
	features := []logistics.Feature{{ExternalID: logistics.ExternalID{Namespace: logistics.ExternalOSMNode, Value: "1"}, Coordinates: coord}, {ExternalID: logistics.ExternalID{Namespace: logistics.ExternalOSMWay, Value: "2"}, Coordinates: coord}}
	p := graph_build_service.Service{Policy: logistics.DefaultFreightPolicy(), Source: fixtureSource{logistics.Dataset{Features: features, References: refs}}}
	g, err := p.Build(context.Background(), logistics.Region{ID: "r", Name: "r", Bounds: logistics.Bounds{West: 29, East: 31, South: 59, North: 61}})

	if err != nil {
		t.Fatal(err)
	}

	if len(g.Nodes) != 1 || len(g.Nodes[0].Modes) != 2 || len(g.Edges) != 0 {
		t.Fatalf("graph=%+v", g)
	}

	if len(g.Coverage.UnroutedNodeModes) != 2 {
		t.Fatalf("missing disconnected warnings: %+v", g.Coverage)
	}
}
