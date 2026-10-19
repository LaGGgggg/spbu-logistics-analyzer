package osmdata

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func TestDownloadIsAtomicAndCached(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1); _, _ = w.Write([]byte("extract")) }))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "extract.osm")
	files := Files{OSM: path, URL: server.URL, SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte("extract")))}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			if err := ensureExtract(context.Background(), files); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if requests.Load() != 1 {
		t.Fatalf("downloaded %d times", requests.Load())
	}
	files.SHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte("wrong")))
	if err := ensureExtract(context.Background(), files); err == nil {
		t.Fatal("checksum mismatch accepted")
	}
	files.OSM = filepath.Join(t.TempDir(), "bad.osm")
	if err := ensureExtract(context.Background(), files); err == nil {
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
	catalog, err := LoadCatalog(path)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Regions["r"].OSM != filepath.Join(dir, "extract.osm") || catalog.Policy.MaximumSnapKM <= 0 {
		t.Fatalf("catalog=%+v", catalog)
	}
	if err := os.WriteFile(path, []byte(`{"regions":{"r":{"osm":"extract.osm"}},"policy":{"maximum_snap_km":-1}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(path); err == nil {
		t.Fatal("negative snap radius accepted")
	}
}
func TestIncompleteWayDoesNotCreateShortcut(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.osm")
	content := `<osm version="0.6"><node id="1" lat="60" lon="30"/><node id="3" lat="60" lon="31"/><way id="10"><nd ref="1"/><nd ref="2"/><nd ref="3"/><tag k="highway" v="primary"/></way></osm>`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	dataset, err := readOSM(context.Background(), path, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(dataset.Tracks) != 0 || len(dataset.Warnings) == 0 {
		t.Fatalf("incomplete way connected: %+v", dataset)
	}
}
