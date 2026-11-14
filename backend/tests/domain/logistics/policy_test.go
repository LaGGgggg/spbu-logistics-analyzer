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
		{"scheduled airport rejected", map[string]string{"aeroway": "aerodrome", "iata": "LED", "scheduled_service": "yes"}, 0, false, ""},
		{"cargo terminal rejected", map[string]string{"aeroway": "cargo_terminal"}, 0, false, ""},
		{"small warehouse rejected", map[string]string{"building": "warehouse"}, 9999, false, ""},
		{"large warehouse rejected", map[string]string{"building": "warehouse"}, 10000, false, ""},
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

func TestOptimizationKeepsWeakCandidatesAtLowLevels(t *testing.T) {

	cases := []struct {
		name  string
		tags  map[string]string
		level int
	}{
		{"passenger station at zero", map[string]string{"railway": "station"}, 0},
		{"passenger station at medium", map[string]string{"railway": "station"}, 50},
		{"yard without purpose at medium", map[string]string{"railway": "yard"}, 50},
		{"harbour without cargo at medium", map[string]string{"harbour": "yes"}, 50},
		{"disused station at zero", map[string]string{"railway": "station", "disused": "yes"}, 0},
		{"fishing port at zero", map[string]string{"port": "fishing"}, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			policy := logistics.FreightPolicy{MaximumSnapKM: 1, Optimization: c.level}
			node, ok := policy.Classify(logistics.Feature{ExternalID: logistics.ExternalID{Namespace: logistics.ExternalOSMNode, Value: "1"}, Tags: c.tags}, nil)

			if !ok || len(node.Modes) == 0 {
				t.Fatalf("rejected at optimization %d", c.level)
			}
		})
	}
}

func TestOptimizationRejectsWeakCandidatesAtHighLevels(t *testing.T) {

	cases := []struct {
		name  string
		tags  map[string]string
		level int
	}{
		{"passenger station", map[string]string{"railway": "station"}, 51},
		{"disused station at medium", map[string]string{"railway": "station", "disused": "yes"}, 50},
		{"fishing port at medium", map[string]string{"port": "fishing"}, 50},
		{"marina at medium", map[string]string{"leisure": "marina"}, 50},
		{"factory never", map[string]string{"industrial": "factory"}, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			policy := logistics.FreightPolicy{MaximumSnapKM: 1, Optimization: c.level}
			_, ok := policy.Classify(logistics.Feature{ExternalID: logistics.ExternalID{Namespace: logistics.ExternalOSMNode, Value: "1"}, Tags: c.tags}, nil)

			if ok {
				t.Fatalf("accepted at optimization %d", c.level)
			}
		})
	}
}

func TestFreightPolicyValidation(t *testing.T) {

	if logistics.DefaultFreightPolicy().Optimization != logistics.MaximumOptimization {
		t.Fatal("default policy must stay strict for backward compatibility")
	}

	for _, level := range []int{-1, 101} {
		if (logistics.FreightPolicy{MaximumSnapKM: 1, Optimization: level}).Validate() == nil {
			t.Fatalf("optimization %d accepted", level)
		}
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
