package osmdata_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/repository/osm"
)

func TestAreaSelectionIgnoresOutsideObjectsAndKeepsCompleteWay(t *testing.T) {

	var xml strings.Builder
	xml.WriteString(`<osm version="0.6">`)

	for id := 100; id < 150; id += 2 {
		fmt.Fprintf(&xml, `<node id="%d" lat="55" lon="25"><tag k="building" v="warehouse"/></node><node id="%d" lat="55.1" lon="25.1"/><way id="%d"><nd ref="%d"/><nd ref="%d"/><tag k="highway" v="primary"/></way><relation id="%d"><member type="way" ref="%d" role="outer"/><tag k="type" v="multipolygon"/><tag k="building" v="warehouse"/></relation>`, id, id+1, id, id, id+1, id, id)
	}

	xml.WriteString(`<node id="1" lat="60" lon="29"/><node id="2" lat="60" lon="30"><tag k="industrial" v="logistics"/><tag k="cargo" v="yes"/></node><node id="3" lat="60" lon="31"/><way id="10"><nd ref="1"/><nd ref="2"/><nd ref="3"/><tag k="highway" v="primary"/><tag k="oneway" v="-1"/></way></osm>`)
	dataset, err := loadSelection(t, xml.String(), 5)

	if err != nil {
		t.Fatal(err)
	}

	if len(dataset.Features) != 1 || dataset.Features[0].ExternalID.Value != "2" || len(dataset.Tracks) != 1 {
		t.Fatalf("outside data was retained: %+v", dataset)
	}

	track := dataset.Tracks[0]

	if !track.OneWay || len(track.Points) != 3 || track.Points[0].ID != "osm:3" || track.Points[2].ID != "osm:1" {
		t.Fatalf("complete way geometry/direction was lost: %+v", track)
	}
}

func TestAreaSelectionRetainsRelationMembersOutsideArea(t *testing.T) {

	xml := `<osm version="0.6">
<relation id="20"><member type="way" ref="10" role="outer"/><member type="way" ref="11" role="outer"/><tag k="type" v="multipolygon"/><tag k="building" v="warehouse"/></relation>
<way id="11"><nd ref="2"/><nd ref="3"/></way>
<way id="10"><nd ref="1"/><nd ref="2"/></way>
<node id="1" lat="60" lon="30"/><node id="2" lat="60" lon="31"/><node id="3" lat="61" lon="31"/>
</osm>`
	dataset, err := loadSelection(t, xml, 5)

	if err != nil {
		t.Fatal(err)
	}

	if len(dataset.Features) != 1 || dataset.Features[0].ExternalID.Namespace != logistics.ExternalOSMRelation || dataset.Features[0].ExternalID.Value != "20" {
		t.Fatalf("relation lost: %+v", dataset)
	}

	if strings.Contains(strings.Join(dataset.Warnings, "\n"), "Incomplete") {
		t.Fatalf("outside relation members lost: %v", dataset.Warnings)
	}
}

func TestAreaSelectionIncludesSnapBufferAndEnforcesLimit(t *testing.T) {

	xml := `<osm version="0.6"><node id="1" lat="60.012" lon="30"/><node id="2" lat="60.013" lon="30"/><way id="10"><nd ref="1"/><nd ref="2"/><tag k="highway" v="primary"/></way></osm>`
	dataset, err := loadSelection(t, xml, 5)

	if err != nil || len(dataset.Tracks) != 1 {
		t.Fatalf("nearby road outside bounds lost: %+v, %v", dataset, err)
	}

	if _, err := loadSelection(t, xml, 1); err == nil || !strings.Contains(err.Error(), "maximum_objects=1") {
		t.Fatalf("selection limit not enforced: %v", err)
	}
}

func loadSelection(t *testing.T, xml string, limit int) (logistics.Dataset, error) {

	t.Helper()
	path := filepath.Join(t.TempDir(), "selection.osm")

	if err := os.WriteFile(path, []byte(xml), 0600); err != nil {
		t.Fatal(err)
	}

	source := osmdata.Source{Catalog: osmdata.Catalog{MaximumObjects: limit, Policy: logistics.FreightPolicy{MaximumSnapKM: 0.5}, Regions: map[string]osmdata.Files{"r": {OSM: path}}}}

	return source.Load(context.Background(), logistics.Region{ID: "r", Bounds: logistics.Bounds{West: 29.99, South: 59.99, East: 30.01, North: 60.01}})
}
