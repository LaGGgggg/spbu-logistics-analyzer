package logistics_test

import (
	"math"
	"slices"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
)

func TestFreightPolicy(t *testing.T) {

	cases := []struct {
		name   string
		tags   map[string]string
		area   float64
		accept bool
		mode   logistics.TransportMode
	}{
		{"passenger station", map[string]string{"railway": "station"}, 0, false, ""},
		{"metro", map[string]string{"railway": "station", "station": "subway", "freight": "yes"}, 0, false, ""},
		{"cargo station", map[string]string{"railway": "station", "freight": "yes"}, 0, true, logistics.TransportRail},
		{"pleasure pier", map[string]string{"man_made": "pier"}, 0, false, ""},
		{"fishing", map[string]string{"port": "fishing", "cargo": "fish"}, 0, false, ""},
		{"river port", map[string]string{"port": "cargo", "port:type": "inland_port"}, 0, true, logistics.TransportRiver},
		{"dry port", map[string]string{"port": "cargo", "port:type": "dryport"}, 0, true, logistics.TransportRail},
		{"scheduled airport insufficient", map[string]string{"aeroway": "aerodrome", "iata": "LED", "scheduled_service": "yes"}, 0, false, ""},
		{"cargo terminal", map[string]string{"aeroway": "cargo_terminal"}, 0, true, logistics.TransportAir},
		{"small warehouse", map[string]string{"building": "warehouse"}, 9999, false, ""},
		{"large warehouse", map[string]string{"building": "warehouse"}, 10000, true, logistics.TransportRoad},
		{"office", map[string]string{"office": "logistics"}, 0, false, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			node, ok := logistics.DefaultFreightPolicy().Classify(logistics.Feature{ExternalID: logistics.ExternalID{Namespace: logistics.ExternalOSMNode, Value: "1"}, Tags: c.tags, AreaM2: c.area}, nil)

			if ok != c.accept {
				t.Fatalf("accepted=%v want %v", ok, c.accept)
			}

			if ok && !slices.Contains(node.Modes, c.mode) {
				t.Fatalf("modes=%v", node.Modes)
			}
		})
	}
}

func TestESRMatching(t *testing.T) {

	for _, v := range []string{"1234-Э", "1234-ПЕР", "1234-ПАР", "123", "1234abc"} {
		if logistics.ESRKey(v) != "" {
			t.Fatalf("service/invalid code %q accepted", v)
		}
	}

	if !logistics.SameExternalID(logistics.ExternalID{Namespace: logistics.ExternalESR, Value: "123456"}, logistics.ExternalID{Namespace: logistics.ExternalESR, Value: "123400"}) {
		t.Fatal("ESR prefixes did not match")
	}
}

func TestRejectNonFiniteValues(t *testing.T) {

	if (logistics.Coordinates{Longitude: math.NaN()}).Validate() == nil {
		t.Fatal("NaN accepted")
	}

	if (logistics.FreightPolicy{MaximumSnapKM: math.Inf(1)}).Validate() == nil {
		t.Fatal("infinite snapping radius accepted")
	}
}
