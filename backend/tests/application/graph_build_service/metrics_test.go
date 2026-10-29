package graph_build_service_test

import (
	"math"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/graph_build_service"
)

func TestCalculateMetrics_EmptyAndSingle(t *testing.T) {
	// Empty graph
	gEmpty := &logistics.LogisticsGraph{}
	graph_build_service.CalculateMetrics(gEmpty)
	if gEmpty.Metrics == nil || len(gEmpty.Metrics.PageRank) != 0 {
		t.Fatalf("expected empty metrics, got %+v", gEmpty.Metrics)
	}

	// Single node
	gSingle := &logistics.LogisticsGraph{
		Nodes: []logistics.LogisticsNode{
			{ID: "n1", Name: "Single Node"},
		},
	}
	graph_build_service.CalculateMetrics(gSingle)
	if gSingle.Metrics.PageRank["n1"] != 1.0 {
		t.Fatalf("expected PageRank 1.0 for single node, got %v", gSingle.Metrics.PageRank["n1"])
	}
	if gSingle.Metrics.Betweenness["n1"] != 0.0 {
		t.Fatalf("expected Betweenness 0.0 for single node, got %v", gSingle.Metrics.Betweenness["n1"])
	}
	if len(gSingle.Metrics.PowerLoad) != 1 || gSingle.Metrics.PowerLoad[0].Degree != 0 {
		t.Fatalf("expected PowerLoad 0 for single node, got %+v", gSingle.Metrics.PowerLoad)
	}
}

func TestCalculateMetrics_PathGraph(t *testing.T) {
	// A -- B -- C (bidirectional path graph)
	// Betweenness of B should be highest (1.0 for normalized (N-1)(N-2) with N=3: (3-1)*(3-2) = 2, paths A->C and C->A go through B, count=2, 2/2 = 1.0)
	g := &logistics.LogisticsGraph{
		Nodes: []logistics.LogisticsNode{
			{ID: "A", Name: "Node A"},
			{ID: "B", Name: "Node B"},
			{ID: "C", Name: "Node C"},
		},
		Edges: []logistics.LogisticsEdge{
			{ID: "e1", FromNodeID: "A", ToNodeID: "B", DistanceKM: 10, Directed: false},
			{ID: "e2", FromNodeID: "B", ToNodeID: "C", DistanceKM: 10, Directed: false},
		},
	}

	graph_build_service.CalculateMetrics(g)

	// Power load: A has degree 1, C has degree 1, B has degree 2
	// Sorted ascending: A and C (degree 1) before B (degree 2)
	pl := g.Metrics.PowerLoad
	if len(pl) != 3 {
		t.Fatalf("expected 3 power load entries, got %d", len(pl))
	}
	if pl[0].Degree != 1 || pl[1].Degree != 1 || pl[2].Degree != 2 {
		t.Fatalf("expected degrees [1, 1, 2], got [%d, %d, %d]", pl[0].Degree, pl[1].Degree, pl[2].Degree)
	}
	if pl[2].NodeID != "B" {
		t.Fatalf("expected node B to have degree 2, got %s", pl[2].NodeID)
	}

	// Betweenness: B should be 1.0, A and C should be 0.0
	if g.Metrics.Betweenness["B"] <= g.Metrics.Betweenness["A"] || g.Metrics.Betweenness["B"] <= g.Metrics.Betweenness["C"] {
		t.Fatalf("node B should have higher betweenness than A and C: %+v", g.Metrics.Betweenness)
	}
	if g.Metrics.Betweenness["B"] != 1.0 {
		t.Fatalf("node B betweenness expected 1.0, got %v", g.Metrics.Betweenness["B"])
	}

	// PageRank: all nodes should have positive rank summing close to 1
	sumPR := 0.0
	for _, pr := range g.Metrics.PageRank {
		sumPR += pr
	}
	if math.Abs(sumPR-1.0) > 0.01 {
		t.Fatalf("PageRank sum should be ~1.0, got %v", sumPR)
	}
	// B has degree 2, so it should have higher PageRank than leaf nodes A and C
	if g.Metrics.PageRank["B"] <= g.Metrics.PageRank["A"] {
		t.Fatalf("node B should have higher PageRank than node A: %+v", g.Metrics.PageRank)
	}

	// Also verify fields were populated on the nodes themselves
	for _, node := range g.Nodes {
		if node.ID == "B" {
			if node.PowerLoad != 2 {
				t.Fatalf("expected node B PowerLoad=2, got %d", node.PowerLoad)
			}
			if node.Betweenness != 1.0 {
				t.Fatalf("expected node B Betweenness=1.0, got %v", node.Betweenness)
			}
		}
	}
}
