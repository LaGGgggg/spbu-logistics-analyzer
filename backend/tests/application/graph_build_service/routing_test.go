package graph_build_service_test

import (
	"context"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/graph_build_service"
)

func TestRailEdgesConnectOnlyAdjacentFacilities(t *testing.T) {

	reference := func(id string, lon, lat float64) logistics.Reference {
		c := logistics.Coordinates{Longitude: lon, Latitude: lat}
		return logistics.Reference{ExternalIDs: []logistics.ExternalID{{Namespace: logistics.ExternalOSMNode, Value: id}}, Modes: []logistics.TransportMode{logistics.TransportRail}, Cargo: true, Source: "test", Coordinates: &c}
	}

	point := func(id string, lon, lat float64) logistics.TrackPoint {
		return logistics.TrackPoint{ID: id, Coordinates: logistics.Coordinates{Longitude: lon, Latitude: lat}}
	}

	mainline := logistics.Track{ID: "osm_way:main", Mode: logistics.TransportRail, Points: []logistics.TrackPoint{
		point("m1", 30.000, 60.000),
		point("m2", 30.010, 60.000),
		point("m3", 30.020, 60.000),
		point("m4", 30.030, 60.000),
		point("m5", 30.040, 60.000),
	}}
	siding := logistics.Track{ID: "osm_way:siding", Mode: logistics.TransportRail, Points: []logistics.TrackPoint{
		point("m3", 30.020, 60.000),
		point("s1", 30.020, 60.003),
	}}

	dataset := logistics.Dataset{
		Tracks: []logistics.Track{mainline, siding},
		References: []logistics.Reference{
			reference("west", 30.000, 60.000),
			reference("middle", 30.020, 60.003),
			reference("east", 30.040, 60.000),
		},
	}

	p := graph_build_service.Service{Policy: logistics.DefaultFreightPolicy(), Source: fixtureSource{dataset}}
	g, err := p.Build(context.Background(), logistics.Region{ID: "r", Name: "r", Bounds: logistics.Bounds{West: 29, East: 31, South: 59, North: 61}})

	if err != nil {
		t.Fatal(err)
	}

	pairs := map[string]bool{}

	for _, edge := range g.Edges {
		pairs[edge.FromNodeID+"->"+edge.ToNodeID] = true
	}

	for _, want := range []string{"osm_node:west->osm_node:middle", "osm_node:middle->osm_node:west", "osm_node:middle->osm_node:east", "osm_node:east->osm_node:middle"} {
		if !pairs[want] {
			t.Errorf("missing direct edge %s; edges=%v", want, pairs)
		}
	}

	if pairs["osm_node:west->osm_node:east"] || pairs["osm_node:east->osm_node:west"] {
		t.Errorf("terminals with an intermediate facility must not link directly; edges=%v", pairs)
	}

	if len(g.Edges) != 4 {
		t.Fatalf("edges=%+v", g.Edges)
	}
}
