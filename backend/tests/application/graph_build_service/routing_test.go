package graph_build_service_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/graph_build_service"
)

func TestRegionalRoutingKeepsDirectionGeometryAndIndependentOrigins(t *testing.T) {

	for _, directed := range []bool{false, true} {
		t.Run(fmt.Sprint("oneway=", directed), func(t *testing.T) {
			points := []logistics.TrackPoint{}
			features := []logistics.Feature{}

			for i := range 5 {
				coordinate := logistics.Coordinates{Longitude: 30 + float64(i)*0.01, Latitude: 60}
				points = append(points, logistics.TrackPoint{ID: fmt.Sprint("p", i), Coordinates: coordinate})

				if i%2 == 0 {
					features = append(features, logistics.Feature{ExternalID: logistics.ExternalID{Namespace: logistics.ExternalOSMNode, Value: fmt.Sprint(i + 1)}, Coordinates: coordinate, Tags: map[string]string{"industrial": "logistics", "cargo": "yes"}})
				}
			}

			builder := graph_build_service.Service{Policy: logistics.DefaultFreightPolicy(), Source: fixtureSource{dataset: logistics.Dataset{Features: features, Tracks: []logistics.Track{{ID: "main", Mode: logistics.TransportRoad, Points: points, OneWay: directed}}}}}
			graph, err := builder.Build(context.Background(), logistics.Region{ID: "r", Name: "r", Bounds: logistics.Bounds{West: 29, South: 59, East: 31, North: 61}})

			if err != nil {
				t.Fatal(err)
			}

			wantEdges := 4

			if directed {
				wantEdges = 2
			}

			if len(graph.Edges) != wantEdges {
				t.Fatalf("origins reused stale distances: %+v", graph.Edges)
			}

			for _, edge := range graph.Edges {
				if len(edge.Geometry) != 3 || edge.DistanceKM <= 0 || edge.Source != "[main]" {
					t.Fatalf("route geometry/source lost: %+v", edge)
				}

				if directed && edge.FromNodeID >= edge.ToNodeID {
					t.Fatalf("oneway ignored: %+v", edge)
				}
			}
		})
	}
}

func TestCompactNetworkRejectsConflictingCoordinates(t *testing.T) {

	coordinate := logistics.Coordinates{Longitude: 30, Latitude: 60}
	feature := logistics.Feature{ExternalID: logistics.ExternalID{Namespace: logistics.ExternalOSMNode, Value: "1"}, Coordinates: coordinate, Tags: map[string]string{"industrial": "logistics", "cargo": "yes"}}
	tracks := []logistics.Track{{ID: "bad", Mode: logistics.TransportRoad, Points: []logistics.TrackPoint{{ID: "same", Coordinates: coordinate}, {ID: "same", Coordinates: logistics.Coordinates{Longitude: 31, Latitude: 60}}}}}
	builder := graph_build_service.Service{Policy: logistics.DefaultFreightPolicy(), Source: fixtureSource{dataset: logistics.Dataset{Features: []logistics.Feature{feature}, Tracks: tracks}}}
	_, err := builder.Build(context.Background(), logistics.Region{ID: "r", Name: "r", Bounds: logistics.Bounds{West: 29, South: 59, East: 32, North: 61}})

	if err == nil || !strings.Contains(err.Error(), "conflicting track point") {
		t.Fatalf("conflicting network identity accepted: %v", err)
	}
}
