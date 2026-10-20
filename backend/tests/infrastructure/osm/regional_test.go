package osmdata_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/repository/osm"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/graph_build_service"
)

func TestRegionalIndexDoesNotCountUnrelatedNodesAgainstGraphLimit(t *testing.T) {

	var xml strings.Builder
	xml.WriteString(`<osm version="0.6">`)

	// Deliberately unordered IDs: binary searches require a sorted index.
	for id := 100; id >= 1; id-- {
		fmt.Fprintf(&xml, `<node id="%d" lat="60" lon="30"><tag k="building" v="yes"/></node>`, id)
	}

	xml.WriteString(`<way id="10"><nd ref="1"/><nd ref="100"/><tag k="highway" v="primary"/></way></osm>`)
	dataset, err := loadSelection(t, xml.String(), 5)

	if err != nil || len(dataset.Tracks) != 1 || len(dataset.Features) != 0 {
		t.Fatalf("unrelated area nodes consumed graph budget: tracks=%d features=%d err=%v", len(dataset.Tracks), len(dataset.Features), err)
	}
}

func TestRegionalIndexHasIndependentLimit(t *testing.T) {

	path := filepath.Join(t.TempDir(), "index.osm")
	xml := `<osm version="0.6"><node id="1" lat="60" lon="30"/><node id="2" lat="60" lon="30"/></osm>`

	if err := os.WriteFile(path, []byte(xml), 0600); err != nil {
		t.Fatal(err)
	}

	source := osmdata.Source{Catalog: osmdata.Catalog{MaximumObjects: 100, MaximumIndexNodes: 1, Regions: map[string]osmdata.Files{"r": {OSM: path}}}}
	_, err := source.Load(context.Background(), logistics.Region{ID: "r", Bounds: logistics.Bounds{West: 29, South: 59, East: 31, North: 61}})

	if !errors.Is(err, graph_build_service.ErrDatasetLimit) || !strings.Contains(err.Error(), "maximum_index_nodes=1") {
		t.Fatalf("independent index guard missing: %v", err)
	}
}

func TestReferenceCanConfirmAnOrdinaryBuilding(t *testing.T) {

	dir := t.TempDir()
	path := filepath.Join(dir, "building.osm")
	xml := `<osm version="0.6"><node id="1" lat="60" lon="30"/><node id="2" lat="60" lon="30.01"/><way id="10"><nd ref="1"/><nd ref="2"/><tag k="building" v="yes"/></way><way id="11"><nd ref="1"/><nd ref="2"/><tag k="building" v="yes"/></way></osm>`

	if err := os.WriteFile(path, []byte(xml), 0600); err != nil {
		t.Fatal(err)
	}

	refs := []logistics.Reference{{ExternalIDs: []logistics.ExternalID{{Namespace: logistics.ExternalOSMWay, Value: "10"}}, Modes: []logistics.TransportMode{logistics.TransportRoad}, Cargo: true, Source: "registry"}}
	payload, err := json.Marshal(refs)

	if err != nil {
		t.Fatal(err)
	}

	refPath := filepath.Join(dir, "references.json")

	if err := os.WriteFile(refPath, payload, 0600); err != nil {
		t.Fatal(err)
	}

	source := osmdata.Source{Catalog: osmdata.Catalog{MaximumObjects: 100, Regions: map[string]osmdata.Files{"r": {OSM: path, References: refPath}}}}
	dataset, err := source.Load(context.Background(), logistics.Region{ID: "r", Bounds: logistics.Bounds{West: 29, South: 59, East: 31, North: 61}})

	if err != nil || len(dataset.Features) != 1 || dataset.Features[0].ExternalID.Value != "10" {
		t.Fatalf("referenced building lost: %+v, %v", dataset.Features, err)
	}
}
