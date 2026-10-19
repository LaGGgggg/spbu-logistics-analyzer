package logistics_service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/logistics_service"
)

type areaRepository struct {
	logistics_service.Repository
	last logistics.Region
}

func (r *areaRepository) RequestAreaBuild(_ context.Context, region logistics.Region) (logistics_service.BuildResult, error) {

	r.last = region

	return logistics_service.BuildResult{JobID: "job"}, nil
}

type areaResolver struct{ version string }

func (r *areaResolver) ResolveArea(context.Context, logistics.Bounds) (logistics_service.AreaSource, error) {
	return logistics_service.AreaSource{ID: "source", Version: r.version}, nil
}

func TestAreaIdentityIncludesBoundsAndSource(t *testing.T) {

	repo := &areaRepository{}
	resolver := &areaResolver{version: "v1"}
	s := logistics_service.NewService(repo, resolver)
	bounds := logistics.Bounds{West: 29, South: 59, East: 31, North: 61}
	first, err := s.RequestAreaBuild(context.Background(), bounds)

	if err != nil {
		t.Fatal(err)
	}

	same, err := s.RequestAreaBuild(context.Background(), bounds)

	if err != nil {
		t.Fatal(err)
	}

	if first.Region.ID != same.Region.ID || first.Region.Bounds != bounds || first.Region.SourceID != "source" {
		t.Fatal("area identity or snapshot mismatch")
	}

	bounds.East = 30.5
	other, err := s.RequestAreaBuild(context.Background(), bounds)

	if err != nil {
		t.Fatal(err)
	}

	if first.Region.ID == other.Region.ID {
		t.Fatal("different bounds share cache identity")
	}

	resolver.version = "v2"
	updated, err := s.RequestAreaBuild(context.Background(), bounds)

	if err != nil {
		t.Fatal(err)
	}

	if updated.Region.ID == other.Region.ID {
		t.Fatal("source versions share cache identity")
	}

	if _, err := s.RequestAreaBuild(context.Background(), logistics.Bounds{}); !errors.Is(err, logistics_service.ErrInvalidInput) {
		t.Fatalf("invalid bounds: %v", err)
	}

	if _, err := logistics_service.NewService(repo).RequestAreaBuild(context.Background(), bounds); !errors.Is(err, logistics_service.ErrAreaUnavailable) {
		t.Fatalf("missing resolver: %v", err)
	}
}
