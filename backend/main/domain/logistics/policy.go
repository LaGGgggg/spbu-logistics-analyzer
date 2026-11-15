package logistics

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// MaximumOptimization is the strictest filtering level: only confirmed freight
// facilities survive, which is the behaviour of the analyzer before the
// optimization degree was introduced.
const MaximumOptimization = 100

type FreightPolicy struct {
	MaximumSnapKM float64 `json:"maximum_snap_km"`
	Optimization  int     `json:"optimization"`

	// MaximumNodes caps the graph size: when classification accepts more
	// facilities, nearby clusters are condensed into single weighted nodes.
	// Zero disables condensation.
	MaximumNodes int `json:"maximum_nodes"`
}

func DefaultFreightPolicy() FreightPolicy {
	return FreightPolicy{MaximumSnapKM: 1, Optimization: MaximumOptimization, MaximumNodes: 2000}
}

func (p FreightPolicy) Validate() error {

	if !inRange(p.MaximumSnapKM, 0.001, 10) || !inRange(float64(p.Optimization), 0, MaximumOptimization) || !inRange(float64(p.MaximumNodes), 0, 1e6) {
		return fmt.Errorf("invalid freight policy")
	}

	return nil
}

// maxTier maps the optimization degree linearly onto the candidate tiers:
// degree 0 accepts every candidate, degree 51 and above only confirmed
// freight facilities.
func (p FreightPolicy) maxTier() int {
	return (MaximumOptimization - p.Optimization) * 2 / MaximumOptimization
}

// candidateTier classifies a feature by the amount of evidence for its freight
// role: 0 — confirmed freight facility, 1 — transport facility without cargo
// confirmation, 2 — explicitly non-freight category.
func candidateTier(t map[string]string) int {

	if t["disused"] == "yes" || t["abandoned"] == "yes" || t["military"] != "" || t["railway"] == "subway" || t["station"] == "subway" || t["port"] == "fishing" || t["port"] == "roro" || t["leisure"] == "marina" {
		return 2
	}

	cargo := t["freight"] == "yes" || (t["cargo"] != "" && t["cargo"] != "no" && t["cargo"] != "passengers")

	if t["railway"] == "container_terminal" || t["industrial"] == "intermodal_freight_terminal" || (t["railway"] == "station" && cargo) || (t["railway"] == "yard" && (cargo || t["railway:yard:purpose"] == "intermodal" || t["railway:yard:purpose"] == "transloading")) {
		return 0
	}

	if t["port"] == "cargo" || (cargo && (t["industrial"] == "port" || t["landuse"] == "port" || t["harbour"] == "yes")) {
		return 0
	}

	return 1
}

// ESRKey matches physical stations by the first four digits. Service records
// (export, transfer, station parks) are deliberately not normalized into stations.
func ESRKey(value string) string {

	if len(value) < 4 || strings.IndexFunc(value, func(r rune) bool {
		return !unicode.IsDigit(r) || r > '9'
	}) >= 0 {
		return ""
	}

	return value[:4]
}

func SameExternalID(a, b ExternalID) bool {

	if a.Namespace != b.Namespace {
		return false
	}

	if a.Namespace == ExternalESR {
		return ESRKey(a.Value) != "" && ESRKey(a.Value) == ESRKey(b.Value)
	}

	return a.Value == b.Value
}

func FeatureIDs(f Feature) []ExternalID {

	ids := []ExternalID{f.ExternalID}

	for key, namespace := range map[string]ExternalNamespace{"esr:user": ExternalESR, "ref:UN:LOCODE": ExternalUNLOCODE, "iata": ExternalIATA, "icao": ExternalICAO} {
		if value := strings.TrimSpace(f.Tags[key]); value != "" {
			ids = append(ids, ExternalID{namespace, value})
		}
	}

	return ids
}

func (p FreightPolicy) Classify(f Feature, references []Reference) (LogisticsNode, bool) {

	t := f.Tags
	tier := candidateTier(t)

	ids := FeatureIDs(f)
	node := LogisticsNode{ID: string(f.ExternalID.Namespace) + ":" + f.ExternalID.Value, Name: t["name"], Coordinates: f.Coordinates, ExternalIDs: ids, Sources: []string{"osm"}, Weight: 1}

	referenceMatched := false

	for _, ref := range references {
		if !ref.Cargo {
			continue
		}

		matched := false

		for _, a := range ids {
			for _, b := range ref.ExternalIDs {
				if SameExternalID(a, b) {
					matched = true
				}
			}
		}

		if !matched {
			continue
		}

		referenceMatched = true
		node.Modes = append(node.Modes, ref.Modes...)
		node.ExternalIDs = append(node.ExternalIDs, ref.ExternalIDs...)
		node.Sources = append(node.Sources, ref.Source)

		if ref.PhysicalID != "" {
			node.ID = "physical:" + ref.PhysicalID
		}

		if node.Name == "" {
			node.Name = ref.Name
		}
	}

	cargo := t["freight"] == "yes" || (t["cargo"] != "" && t["cargo"] != "no" && t["cargo"] != "passengers")

	if t["railway"] == "container_terminal" || t["industrial"] == "intermodal_freight_terminal" || (t["railway"] == "station" && cargo) || (t["railway"] == "yard" && (cargo || t["railway:yard:purpose"] == "intermodal" || t["railway:yard:purpose"] == "transloading")) {
		node.Modes = append(node.Modes, TransportRail)
		node.Kind = "freight_terminal"
	}

	port := t["port"] == "cargo" || (cargo && (t["industrial"] == "port" || t["landuse"] == "port" || t["harbour"] == "yes"))

	if port {
		node.Modes = append(node.Modes, portMode(t))
		node.Kind = "port"
	}

	if len(node.Modes) == 0 && p.maxTier() >= 1 {
		if t["railway"] != "" && t["railway"] != "rail" {
			node.Modes = append(node.Modes, TransportRail)
		}

		if t["industrial"] == "port" || t["landuse"] == "port" || t["harbour"] == "yes" || (t["port"] != "" && t["port"] != "cargo") {
			node.Modes = append(node.Modes, portMode(t))

			if node.Kind == "" {
				node.Kind = "port"
			}
		}
	}

	if referenceMatched {
		tier = 0
	}

	if len(node.Modes) == 0 || tier > p.maxTier() {
		return LogisticsNode{}, false
	}

	slices.Sort(node.Modes)
	node.Modes = slices.Compact(node.Modes)
	slices.SortFunc(node.ExternalIDs, func(a, b ExternalID) int {
		return strings.Compare(string(a.Namespace)+":"+a.Value, string(b.Namespace)+":"+b.Value)
	})
	node.ExternalIDs = slices.Compact(node.ExternalIDs)
	slices.Sort(node.Sources)
	node.Sources = slices.Compact(node.Sources)

	if node.Name == "" {
		node.Name = node.ID
	}

	if node.Kind == "" {
		node.Kind = "freight_facility"
	}

	node.Cargo = t["cargo"]

	return node, true
}

func portMode(t map[string]string) TransportMode {

	if t["port:type"] == "inland_port" {
		return TransportRiver
	}

	if t["port:type"] == "dryport" {
		return TransportRail
	}

	return TransportSea
}
