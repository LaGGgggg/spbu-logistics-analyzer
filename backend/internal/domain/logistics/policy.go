package logistics

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

type FreightPolicy struct {
	MinimumWarehouseAreaM2 float64 `json:"minimum_warehouse_area_m2"`
	MaximumSnapKM          float64 `json:"maximum_snap_km"`
}

func DefaultFreightPolicy() FreightPolicy {
	return FreightPolicy{MinimumWarehouseAreaM2: 10000, MaximumSnapKM: 1}
}

func (p FreightPolicy) Validate() error {

	if !inRange(p.MinimumWarehouseAreaM2, 0, 1e12) || !inRange(p.MaximumSnapKM, 0.001, 10) {
		return fmt.Errorf("invalid freight policy")
	}

	return nil
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

	if t["disused"] == "yes" || t["abandoned"] == "yes" || t["military"] != "" || t["railway"] == "subway" || t["station"] == "subway" || t["port"] == "fishing" || t["port"] == "roro" || t["leisure"] == "marina" {
		return LogisticsNode{}, false
	}

	ids := FeatureIDs(f)
	node := LogisticsNode{ID: string(f.ExternalID.Namespace) + ":" + f.ExternalID.Value, Name: t["name"], Coordinates: f.Coordinates, ExternalIDs: ids, Sources: []string{"osm"}}

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
		node.Modes = append(node.Modes, TransportRail, TransportRoad)
		node.Kind = "freight_terminal"
	}

	port := t["port"] == "cargo" || (cargo && (t["industrial"] == "port" || t["landuse"] == "port" || t["harbour"] == "yes"))

	if port {
		mode := TransportSea

		if t["port:type"] == "inland_port" {
			mode = TransportRiver
		}

		if t["port:type"] == "dryport" {
			mode = TransportRail
		}

		node.Modes = append(node.Modes, mode, TransportRoad)
		node.Kind = "port"
	}

	if t["aeroway"] == "cargo" || t["aeroway"] == "cargo_terminal" || (cargo && (t["aeroway"] == "aerodrome" || t["aeroway"] == "terminal")) {
		node.Modes = append(node.Modes, TransportAir, TransportRoad)
		node.Kind = "cargo_airport"
	}

	warehouse := t["building"] == "warehouse" || t["industrial"] == "warehouse" || t["industrial"] == "logistics" || t["man_made"] == "distribution_centre"

	if warehouse && (cargo || t["man_made"] == "distribution_centre" || f.AreaM2 >= p.MinimumWarehouseAreaM2) {
		node.Modes = append(node.Modes, TransportRoad)

		if node.Kind == "" {
			node.Kind = "warehouse"
		}
	}

	if len(node.Modes) == 0 {
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
