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
	"github.com/paulmach/osm/osmxml"
)

func scan(ctx context.Context, path string, visit func(osm.Object) error) error {

	f, err := os.Open(path)

	if err != nil {
		return err
	}

	defer f.Close()

	var scanner osm.Scanner

	if strings.HasSuffix(path, ".pbf") {
		scanner, err = NewPBFScanner(ctx, f, 2)
	} else {
		scanner = osmxml.New(ctx, f)
	}

	if err != nil {
		return err
	}

	defer scanner.Close()

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := visit(scanner.Object()); err != nil {
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

	if t["railway"] != "" && t["railway"] != "rail" {
		return true
	}

	for _, key := range []string{"aeroway", "industrial", "port", "cargo", "harbour"} {
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

	if t["waterway"] == "river" || t["waterway"] == "canal" {
		if t["boat"] == "yes" || t["ship"] == "yes" || t["motorboat"] == "yes" {
			return logistics.TransportRiver, true
		}
	}

	return "", false
}

func readOSM(ctx context.Context, path string, limit int) (logistics.Dataset, error) {

	result := logistics.Dataset{}
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
	err := scan(ctx, path, func(obj osm.Object) error {
		switch o := obj.(type) {
		case *osm.Way:
			t := tagsMap(o.Tags)
			_, network := trackMode(t)

			if network || candidate(t) {
				if err := check(); err != nil {
					return err
				}

				ways[o.ID] = o
			}
		case *osm.Relation:
			if candidate(tagsMap(o.Tags)) && (o.Tags.Find("type") == "multipolygon" || o.Tags.Find("type") == "site") {
				if err := check(); err != nil {
					return err
				}

				relations = append(relations, o)

				for _, member := range o.Members {
					if member.Type == osm.TypeWay && member.Role != "inner" {
						wantedWays[osm.WayID(member.Ref)] = true
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return result, err
	}

	if len(wantedWays) > 0 {
		if err := scan(ctx, path, func(obj osm.Object) error {
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
		}
	}

	if len(wantedNodes) > limit {
		return result, fmt.Errorf("OSM node selection exceeds maximum_objects=%d", limit)
	}

	err = scan(ctx, path, func(obj osm.Object) error {
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

		if candidate(t) {
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

			directed := oneway == "yes" || oneway == "1" || oneway == "true" || oneway == "-1"
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
