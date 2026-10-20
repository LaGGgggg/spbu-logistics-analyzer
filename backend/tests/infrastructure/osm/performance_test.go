package osmdata_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/repository/osm"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/graph_build_service"
)

type measuredSource struct {
	source  osmdata.Source
	elapsed time.Duration
}

func (s *measuredSource) Load(ctx context.Context, region logistics.Region) (logistics.Dataset, error) {

	start := time.Now()
	dataset, err := s.source.Load(ctx, region)
	s.elapsed += time.Since(start)

	return dataset, err
}

// Opt in with OSM_BENCH_FILE pointing to a regional extract; never downloads data.
func BenchmarkRegionalGraphBuild(b *testing.B) {

	path := os.Getenv("OSM_BENCH_FILE")

	if path == "" {
		b.Skip("set OSM_BENCH_FILE to benchmark a regional PBF/XML extract")
	}

	policy := logistics.DefaultFreightPolicy()
	policy.MaximumSnapKM = 0.5
	source := &measuredSource{source: osmdata.Source{Catalog: osmdata.Catalog{MaximumObjects: 5000000, Policy: policy, Regions: map[string]osmdata.Files{"benchmark": {OSM: path}}}}}
	builder := graph_build_service.Service{Source: source, Policy: policy}
	region := logistics.Region{ID: "benchmark", Name: "South Saint Petersburg", Bounds: logistics.Bounds{West: 30.10811, South: 59.80699, East: 30.48226, North: 59.93393}}
	if os.Getenv("OSM_BENCH_REGION") == "spb-leningrad" {
		region.Bounds = logistics.Bounds{West: 27.4, South: 58.4, East: 35.7, North: 61.4}
		region.Name = "SPb and Leningrad region bounding box"
	}
	var total time.Duration

	b.ReportAllocs()

	for b.Loop() {
		start := time.Now()
		graph, err := builder.Build(context.Background(), region)
		total += time.Since(start)

		if err != nil {
			b.Fatal(err)
		}

		b.ReportMetric(float64(len(graph.Nodes)), "nodes")
		b.ReportMetric(float64(len(graph.Edges)), "edges")
	}

	b.ReportMetric(float64(source.elapsed.Milliseconds())/float64(b.N), "osm_ms/op")
	b.ReportMetric(float64((total-source.elapsed).Milliseconds())/float64(b.N), "graph_ms/op")
}
