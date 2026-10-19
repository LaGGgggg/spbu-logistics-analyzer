package osmdata_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/repository/osm"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/tests/testutil"
)

func TestDownloadIsAtomicAndCached(t *testing.T) {

	fixture, err := os.ReadFile(testutil.FixturePath(t, "freight.osm"))

	if err != nil {
		t.Fatal(err)
	}

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = w.Write(fixture)
	}))

	defer server.Close()

	path := filepath.Join(t.TempDir(), "extract.osm")
	files := osmdata.Files{OSM: path, URL: server.URL, SHA256: fmt.Sprintf("%x", sha256.Sum256(fixture))}
	region := logistics.Region{ID: "r"}
	load := func(files osmdata.Files) error {
		source := osmdata.Source{Catalog: osmdata.Catalog{MaximumObjects: 1000, Regions: map[string]osmdata.Files{"r": files}}}
		_, err := source.Load(context.Background(), region)

		return err
	}
	var wg sync.WaitGroup

	for range 8 {
		wg.Go(func() {
			if err := load(files); err != nil {
				t.Error(err)
			}
		})
	}

	wg.Wait()

	if requests.Load() != 1 {
		t.Fatalf("downloaded %d times", requests.Load())
	}

	files.SHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte("wrong")))

	if err := load(files); err == nil {
		t.Fatal("checksum mismatch accepted")
	}

	files.OSM = filepath.Join(t.TempDir(), "bad.osm")

	if err := load(files); err == nil {
		t.Fatal("bad download published")
	}

	if _, err := os.Stat(files.OSM); !os.IsNotExist(err) {
		t.Fatal("partial download is visible")
	}
}

func TestCatalogResolvesPathsAndRejectsInvalidPolicy(t *testing.T) {

	dir := t.TempDir()
	path := filepath.Join(dir, "sources.json")

	if err := os.WriteFile(path, []byte(`{"regions":{"r":{"osm":"extract.osm"}}}`), 0600); err != nil {
		t.Fatal(err)
	}

	catalog, err := osmdata.LoadCatalog(path)

	if err != nil {
		t.Fatal(err)
	}

	if catalog.Regions["r"].OSM != filepath.Join(dir, "extract.osm") || catalog.Policy.MaximumSnapKM <= 0 {
		t.Fatalf("catalog=%+v", catalog)
	}

	if err := os.WriteFile(path, []byte(`{"regions":{"r":{"osm":"extract.osm"}},"policy":{"maximum_snap_km":-1}}`), 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := osmdata.LoadCatalog(path); err == nil {
		t.Fatal("negative snap radius accepted")
	}
}

func TestIncompleteWayDoesNotCreateShortcut(t *testing.T) {

	path := filepath.Join(t.TempDir(), "bad.osm")
	content := `<osm version="0.6"><node id="1" lat="60" lon="30"/><node id="3" lat="60" lon="31"/><way id="10"><nd ref="1"/><nd ref="2"/><nd ref="3"/><tag k="highway" v="primary"/></way></osm>`

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	source := osmdata.Source{Catalog: osmdata.Catalog{MaximumObjects: 100, Regions: map[string]osmdata.Files{"r": {OSM: path}}}}
	dataset, err := source.Load(context.Background(), logistics.Region{ID: "r"})

	if err != nil {
		t.Fatal(err)
	}

	if len(dataset.Tracks) != 0 || !strings.Contains(strings.Join(dataset.Warnings, "\n"), "Incomplete OSM way 10") {
		t.Fatalf("incomplete way connected: %+v", dataset)
	}
}
