package osmdata

import (
	"context"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/paulmach/osm/osmxml"
)

type objectKinds uint8

const (
	nodeObjects objectKinds = 1 << iota
	wayObjects
	relationObjects
)

func scan(ctx context.Context, path string, kinds objectKinds, keep func(osm.Object) bool, visit func(osm.Object) error) error {

	f, err := os.Open(path)

	if err != nil {
		return err
	}

	defer f.Close()

	var scanner osm.Scanner

	if strings.HasSuffix(path, ".pbf") {
		pbf := osmpbf.New(ctx, f, 2)
		pbf.SkipNodes = kinds&nodeObjects == 0
		pbf.SkipWays = kinds&wayObjects == 0
		pbf.SkipRelations = kinds&relationObjects == 0

		// PBF filters run on decoder goroutines. Their captured state must be
		// immutable during this pass; rejected objects reuse decoder memory.
		if keep != nil {
			pbf.FilterNode = func(node *osm.Node) bool { return keep(node) }
			pbf.FilterWay = func(way *osm.Way) bool { return keep(way) }
			pbf.FilterRelation = func(relation *osm.Relation) bool { return keep(relation) }
		}

		scanner = pbf
	} else {
		scanner = osmxml.New(ctx, f)
	}

	defer scanner.Close()

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}

		object := scanner.Object()

		switch object.(type) {
		case *osm.Node:
			if kinds&nodeObjects == 0 {
				continue
			}
		case *osm.Way:
			if kinds&wayObjects == 0 {
				continue
			}
		case *osm.Relation:
			if kinds&relationObjects == 0 {
				continue
			}
		}

		if keep != nil && !keep(object) {
			continue
		}

		if err := visit(object); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func tagsMap(tags osm.Tags) map[string]string {

	m := make(map[string]string, len(tags))

	for _, t := range tags {
		m[t.Key] = t.Value
	}

	return m
}

func candidate(t map[string]string) bool {

	for _, key := range []string{"railway", "aeroway", "industrial", "building", "port", "cargo", "harbour", "man_made"} {
		if t[key] != "" {
			return true
		}
	}

	return false
}

func trackMode(t map[string]string) (logistics.TransportMode, bool) {

	if t["access"] == "no" || t["access"] == "private" || t["hgv"] == "no" || t["disused"] == "yes" || t["abandoned"] == "yes" {
		return "", false
	}

	if t["railway"] == "rail" && t["railway:traffic_mode"] != "passenger" {
		return logistics.TransportRail, true
	}

	if slices.Contains([]string{"motorway", "motorway_link", "trunk", "trunk_link", "primary", "primary_link", "secondary", "secondary_link", "tertiary", "tertiary_link", "unclassified", "residential", "service", "living_street"}, t["highway"]) {
		if t["motor_vehicle"] == "no" || t["motorcar"] == "no" {
			return "", false
		}

		return logistics.TransportRoad, true
	}

	if t["waterway"] == "river" || t["waterway"] == "canal" {
		if t["boat"] == "yes" || t["ship"] == "yes" || t["motorboat"] == "yes" {
			return logistics.TransportRiver, true
		}
	}

	return "", false
}

func readOSM(ctx context.Context, path string, limit int, bounds logistics.Bounds, snapKM float64) (logistics.Dataset, error) {

	result := logistics.Dataset{}
	selectionBounds := bufferedBounds(bounds, snapKM)
	localNodes, err := selectAreaNodes(ctx, path, selectionBounds, limit)

	if err != nil {
		return result, err
	}

	ways := map[osm.WayID]*osm.Way{}
	relations := []*osm.Relation{}
	wantedWays := map[osm.WayID]bool{}
	wantedNodes := map[osm.NodeID]bool{}
	points := map[osm.NodeID]logistics.Coordinates{}
	count := 0
	check := func() error {
		count++

		if count > limit {
			return fmt.Errorf("OSM selection exceeds maximum_objects=%d", limit)
		}

		return nil
	}
	err = scan(ctx, path, wayObjects, func(obj osm.Object) bool {
		way, ok := obj.(*osm.Way)

		return ok && touchesArea(way, localNodes)
	}, func(obj osm.Object) error {
		if o, ok := obj.(*osm.Way); ok {
			if touchesArea(o, localNodes) {
				if err := check(); err != nil {
					return err
				}

				ways[o.ID] = o
			}
		}

		return nil
	})

	if err != nil {
		return result, err
	}

	// Select relations after ways so selection also works for unordered XML.
	err = scan(ctx, path, relationObjects, nil, func(obj osm.Object) error {
		o, ok := obj.(*osm.Relation)

		if !ok || !candidate(tagsMap(o.Tags)) || (o.Tags.Find("type") != "multipolygon" && o.Tags.Find("type") != "site") {
			return nil
		}

		selected := false

		for _, member := range o.Members {
			if member.Type == osm.TypeWay && member.Role != "inner" && ways[osm.WayID(member.Ref)] != nil {
				selected = true
				break
			}
		}

		if !selected {
			return nil
		}

		if err := check(); err != nil {
			return err
		}

		relations = append(relations, o)

		for _, member := range o.Members {
			if member.Type == osm.TypeWay && member.Role != "inner" {
				wantedWays[osm.WayID(member.Ref)] = true
			}
		}

		return nil
	})

	if err != nil {
		return result, err
	}

	for id, way := range ways {
		t := tagsMap(way.Tags)
		_, network := trackMode(t)

		if !network && !candidate(t) && !wantedWays[id] {
			delete(ways, id)
		}
	}

	missingWays := map[osm.WayID]bool{}

	for id := range wantedWays {
		if ways[id] == nil {
			missingWays[id] = true
		}
	}

	if len(missingWays) > 0 {
		if err := scan(ctx, path, wayObjects, func(obj osm.Object) bool {
			way, ok := obj.(*osm.Way)

			return ok && missingWays[way.ID]
		}, func(obj osm.Object) error {
			if way, ok := obj.(*osm.Way); ok && wantedWays[way.ID] && ways[way.ID] == nil {
				if err := check(); err != nil {
					return err
				}

				ways[way.ID] = way
			}

			return nil
		}); err != nil {
			return result, err
		}
	}

	for _, way := range ways {
		for _, nd := range way.Nodes {
			wantedNodes[nd.ID] = true

			if len(wantedNodes) > limit {
				return result, fmt.Errorf("OSM node selection exceeds maximum_objects=%d", limit)
			}
		}
	}

	// Release the spatial index before loading complete way geometries.
	localNodes = nil

	err = scan(ctx, path, nodeObjects, func(obj osm.Object) bool {
		node, ok := obj.(*osm.Node)

		return ok && (wantedNodes[node.ID] || inside(selectionBounds, logistics.Coordinates{Longitude: node.Lon, Latitude: node.Lat}))
	}, func(obj osm.Object) error {
		node, ok := obj.(*osm.Node)

		if !ok {
			return nil
		}

		coordinate := logistics.Coordinates{Longitude: node.Lon, Latitude: node.Lat}

		if err := coordinate.Validate(); err != nil {
			return err
		}

		if wantedNodes[node.ID] {
			points[node.ID] = coordinate
		}

		t := tagsMap(node.Tags)

		if inside(selectionBounds, coordinate) && candidate(t) {
			if err := check(); err != nil {
				return err
			}

			result.Features = append(result.Features, logistics.Feature{ExternalID: logistics.ExternalID{Namespace: logistics.ExternalOSMNode, Value: strconv.FormatInt(int64(node.ID), 10)}, Coordinates: coordinate, Tags: t})
		}

		return nil
	})

	if err != nil {
		return result, err
	}

	ids := make([]osm.WayID, 0, len(ways))

	for id := range ways {
		ids = append(ids, id)
	}

	slices.Sort(ids)

	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		way := ways[id]
		t := tagsMap(way.Tags)
		geometry := []logistics.TrackPoint{}
		complete := true

		for _, nd := range way.Nodes {
			c, ok := points[nd.ID]

			if !ok {
				complete = false
				break
			}

			geometry = append(geometry, logistics.TrackPoint{ID: "osm:" + strconv.FormatInt(int64(nd.ID), 10), Coordinates: c})
		}

		if !complete || len(geometry) < 2 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Incomplete OSM way %d skipped", id))
			continue
		}

		if mode, ok := trackMode(t); ok {
			oneway := t["oneway"]

			if mode == logistics.TransportRail && t["oneway:railway"] != "" {
				oneway = t["oneway:railway"]
			}

			if oneway == "-1" {
				slices.Reverse(geometry)
			}

			directed := oneway == "yes" || oneway == "1" || oneway == "true" || oneway == "-1" || (mode == logistics.TransportRoad && t["junction"] == "roundabout" && oneway != "no")
			result.Tracks = append(result.Tracks, logistics.Track{ID: fmt.Sprintf("osm_way:%d", id), Mode: mode, Points: geometry, OneWay: directed})
		}

		if candidate(t) {
			result.Features = append(result.Features, logistics.Feature{ExternalID: logistics.ExternalID{Namespace: logistics.ExternalOSMWay, Value: strconv.FormatInt(int64(id), 10)}, Coordinates: center(geometry), Tags: t, AreaM2: area(geometry)})
		}
	}

	for _, rel := range relations {
		geometry := []logistics.TrackPoint{}
		complete := true

		for _, member := range rel.Members {
			if member.Type != osm.TypeWay || member.Role == "inner" {
				continue
			}

			way := ways[osm.WayID(member.Ref)]

			if way == nil {
				complete = false
				break
			}

			for _, nd := range way.Nodes {
				c, ok := points[nd.ID]

				if !ok {
					complete = false
					break
				}

				geometry = append(geometry, logistics.TrackPoint{Coordinates: c})
			}
		}

		if !complete || len(geometry) == 0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Incomplete OSM relation %d skipped", rel.ID))
			continue
		}

		result.Features = append(result.Features, logistics.Feature{ExternalID: logistics.ExternalID{Namespace: logistics.ExternalOSMRelation, Value: strconv.FormatInt(int64(rel.ID), 10)}, Coordinates: center(geometry), Tags: tagsMap(rel.Tags)})
	}

	result.Warnings = append(result.Warnings, "OSM network limited to ways with a node in the selected bounds and snap buffer; routes leaving this network may be unavailable")

	return result, nil
}

func center(points []logistics.TrackPoint) logistics.Coordinates {

	var c logistics.Coordinates

	for _, p := range points {
		c.Longitude += p.Coordinates.Longitude
		c.Latitude += p.Coordinates.Latitude
	}

	c.Longitude /= float64(len(points))
	c.Latitude /= float64(len(points))

	return c
}

func area(points []logistics.TrackPoint) float64 {

	if len(points) < 4 || points[0].ID != points[len(points)-1].ID {
		return 0
	}

	c := center(points)
	scale := math.Cos(c.Latitude * math.Pi / 180)
	sum := 0.0

	for i := 1; i < len(points); i++ {
		a, b := points[i-1].Coordinates, points[i].Coordinates
		sum += (a.Longitude-c.Longitude)*(b.Latitude-c.Latitude) - (b.Longitude-c.Longitude)*(a.Latitude-c.Latitude)
	}

	return math.Abs(sum) * 0.5 * 111195.08 * 111195.08 * scale
}
