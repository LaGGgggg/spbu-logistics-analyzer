package graph_build_service_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/graph_build_service"
)

func TestLongRouteSimplificationKeepsCornersEndpointsAndOriginalDistance(t *testing.T) {

	points := []logistics.TrackPoint{}

	for i := range 1001 {
		coordinate := logistics.Coordinates{Longitude: 30 + float64(i)*0.0001, Latitude: 60}

		if i > 500 {
			coordinate = logistics.Coordinates{Longitude: 30.05, Latitude: 60 + float64(i-500)*0.0001}
		}

		points = append(points, logistics.TrackPoint{ID: fmt.Sprint(i), Coordinates: coordinate})
	}

	features := []logistics.Feature{}

	for _, i := range []int{0, 1000} {
		features = append(features, logistics.Feature{ExternalID: logistics.ExternalID{Namespace: logistics.ExternalOSMNode, Value: fmt.Sprint(i + 1)}, Coordinates: points[i].Coordinates, Tags: map[string]string{"industrial": "logistics", "cargo": "yes"}})
	}

	source := fixtureSource{dataset: logistics.Dataset{Features: features, Tracks: []logistics.Track{{ID: "bend", Mode: logistics.TransportRoad, Points: points}}}}
	region := logistics.Region{ID: "r", Name: "r", Bounds: logistics.Bounds{West: 29, South: 59, East: 31, North: 61}}
	policy := logistics.DefaultFreightPolicy()
	policy.RouteGeometryToleranceM = 0
	exact, err := (graph_build_service.Service{Source: source, Policy: policy}).Build(context.Background(), region)

	if err != nil || len(exact.Edges) != 2 || len(exact.Edges[0].Geometry) != len(points) {
		t.Fatalf("exact geometry mode broken: edges=%d err=%v", len(exact.Edges), err)
	}

	policy.RouteGeometryToleranceM = 10
	compact, err := (graph_build_service.Service{Source: source, Policy: policy}).Build(context.Background(), region)

	if err != nil || len(compact.Edges) != 2 {
		t.Fatalf("simplified route lost: edges=%d err=%v", len(compact.Edges), err)
	}

	for i, edge := range compact.Edges {
		if edge.DistanceKM != exact.Edges[i].DistanceKM || edge.Source != exact.Edges[i].Source {
			t.Fatalf("simplification changed routing result: %+v", edge)
		}

		geometry := edge.Geometry

		if len(geometry) != 3 || geometry[1] != points[500].Coordinates || geometry[0] != exact.Edges[i].Geometry[0] || geometry[2] != exact.Edges[i].Geometry[1000] {
			t.Fatalf("route corner/endpoints lost: %+v", geometry)
		}
	}

	if !strings.Contains(strings.Join(compact.Coverage.Warnings, "\n"), "simplified") {
		t.Fatal("geometry simplification not reported in coverage")
	}
}
