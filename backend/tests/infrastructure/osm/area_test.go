package osmdata_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/application/logistics_service"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/repository/osm"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/tests/testutil"
)

func TestAreaSelectsSmallestCoveringSource(t *testing.T) {
	large := logistics.Bounds{West: 20, South: 50, East: 40, North: 70}
	small := logistics.Bounds{West: 29, South: 59, East: 31, North: 61}
	s := osmdata.Source{Catalog: osmdata.Catalog{Policy: logistics.DefaultFreightPolicy(), Regions: map[string]osmdata.Files{"large": {Bounds: &large}, "small": {Bounds: &small}}}}
	r, err := s.ResolveArea(context.Background(), small)
	if err != nil || r.ID != "small" {
		t.Fatalf("source=%+v %v", r, err)
	}
	outside := logistics.Bounds{West: 1, South: 1, East: 2, North: 2}
	if _, err := s.ResolveArea(context.Background(), outside); !errors.Is(err, logistics_service.ErrAreaUnavailable) {
		t.Fatalf("outside coverage: %v", err)
	}
	s.Catalog.OverpassURL = "http://example.invalid"
	r, err = s.ResolveArea(context.Background(), outside)
	if err != nil || r.ID == "" || r.ID == "small" || r.ID == "large" {
		t.Fatalf("fallback=%+v %v", r, err)
	}
}
func TestOverpassLoadsArbitraryAreaAndCachesXML(t *testing.T) {
	fixture, err := os.ReadFile(testutil.FixturePath(t, "freight.osm"))
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != "POST" {
			t.Error("expected POST")
		}
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		query := r.Form.Get("data")
		if !strings.Contains(query, "(59,29,61,31)") || !strings.Contains(query, "(._;>>;)") {
			t.Errorf("query=%s", query)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()
	s := osmdata.Source{Catalog: osmdata.Catalog{OverpassURL: server.URL, CacheDirectory: t.TempDir(), MaximumObjects: 1000}}
	bounds := logistics.Bounds{West: 29, South: 59, East: 31, North: 61}
	selected, err := s.ResolveArea(context.Background(), bounds)
	if err != nil {
		t.Fatal(err)
	}
	region := logistics.Region{ID: "area-test", SourceID: selected.ID, SourceVersion: selected.Version, Name: "User input", Bounds: bounds}
	for range 2 {
		dataset, err := s.Load(context.Background(), region)
		if err != nil {
			t.Fatal(err)
		}
		if len(dataset.Tracks) != 1 || len(dataset.Features) != 4 {
			t.Fatalf("dataset=%+v", dataset)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("requests=%d", calls.Load())
	}
}
func TestOverpassRuntimeErrorIsNotCached(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<osm version="0.6"><remark>runtime error: timeout</remark></osm>`))
	}))
	defer server.Close()
	cache := t.TempDir()
	s := osmdata.Source{Catalog: osmdata.Catalog{OverpassURL: server.URL, CacheDirectory: cache}}
	bounds := logistics.Bounds{West: 1, South: 1, East: 2, North: 2}
	selected, err := s.ResolveArea(context.Background(), bounds)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Load(context.Background(), logistics.Region{ID: "area-test", SourceID: selected.ID, SourceVersion: selected.Version, Bounds: bounds})
	if err == nil {
		t.Fatal("runtime error accepted")
	}
	entries, err := os.ReadDir(cache)
	if err != nil || len(entries) != 0 {
		t.Fatalf("invalid extract cached: %v %v", entries, err)
	}
}

func TestChangedAreaSourceIsNotUsedSilently(t *testing.T) {
	bounds := logistics.Bounds{West: 29, South: 59, East: 31, North: 61}
	s := osmdata.Source{Catalog: osmdata.Catalog{OverpassURL: "https://example.invalid/interpreter", Policy: logistics.DefaultFreightPolicy()}}
	version, err := s.ResolveArea(context.Background(), bounds)
	if err != nil {
		t.Fatal(err)
	}
	region := logistics.Region{ID: "area-test", SourceID: version.ID, SourceVersion: version.Version, Bounds: bounds}
	s.Catalog.Policy.MaximumSnapKM = 2
	if _, err := s.Load(context.Background(), region); err == nil || !strings.Contains(err.Error(), "configuration changed") {
		t.Fatalf("changed source accepted: %v", err)
	}
}
