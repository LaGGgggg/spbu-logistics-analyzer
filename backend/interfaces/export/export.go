package export

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/logistics"
)

var Formats = []string{"nodes.csv", "edges.csv", "adjacency.csv", "incidence.csv", "geojson", "graphml"}

func ContentType(format string) string {
	switch format {
	case "geojson":
		return "application/geo+json"
	case "graphml":
		return "application/graphml+xml"
	default:
		return "text/csv; charset=utf-8"
	}
}

func Write(w io.Writer, g logistics.LogisticsGraph, format string) error {

	if !slices.Contains(Formats, format) {
		return fmt.Errorf("unknown export format")
	}

	if format == "geojson" {
		return geoJSON(w, g)
	}

	if format == "graphml" {
		return graphML(w, g)
	}

	c := csv.NewWriter(w)
	row := func(values ...string) error {
		return c.Write(values)
	}
	number := func(n float64) string {
		return strconv.FormatFloat(n, 'g', -1, 64)
	}

	switch format {
	case "nodes.csv":
		if err := row("id", "name", "region_id", "longitude", "latitude", "modes", "external_ids", "kind", "cargo", "sources"); err != nil {
			return err
		}

		for _, n := range g.Nodes {
			modes := []string{}

			for _, m := range n.Modes {
				modes = append(modes, string(m))
			}

			ids, _ := json.Marshal(n.ExternalIDs)

			if err := row(n.ID, safeCell(n.Name), n.RegionID, number(n.Coordinates.Longitude), number(n.Coordinates.Latitude), strings.Join(modes, ";"), string(ids), n.Kind, safeCell(n.Cargo), safeCell(strings.Join(n.Sources, ";"))); err != nil {
				return err
			}
		}
	case "edges.csv":
		if err := row("id", "from_node_id", "to_node_id", "mode", "distance_km", "directed", "geometry", "source"); err != nil {
			return err
		}

		for _, e := range g.Edges {
			geometry, _ := json.Marshal(e.Geometry)

			if err := row(e.ID, e.FromNodeID, e.ToNodeID, string(e.Mode), number(e.DistanceKM), strconv.FormatBool(e.Directed), string(geometry), safeCell(e.Source)); err != nil {
				return err
			}
		}
	case "adjacency.csv":
		header := []string{"node_id"}

		for _, n := range g.Nodes {
			header = append(header, n.ID)
		}

		if err := c.Write(header); err != nil {
			return err
		}

		values := map[string]map[string]float64{}
		put := func(from, to string, d float64) {
			if values[from] == nil {
				values[from] = map[string]float64{}
			}

			old, ok := values[from][to]

			if !ok || d < old {
				values[from][to] = d
			}
		}

		for _, e := range g.Edges {
			put(e.FromNodeID, e.ToNodeID, e.DistanceKM)

			if !e.Directed {
				put(e.ToNodeID, e.FromNodeID, e.DistanceKM)
			}
		}

		for _, n := range g.Nodes {
			r := []string{n.ID}

			for _, other := range g.Nodes {
				value := ""

				if d, ok := values[n.ID][other.ID]; ok {
					value = number(d)
				}

				r = append(r, value)
			}

			if err := c.Write(r); err != nil {
				return err
			}
		}
	case "incidence.csv":
		header := []string{"node_id"}

		for _, e := range g.Edges {
			header = append(header, e.ID)
		}

		if err := c.Write(header); err != nil {
			return err
		}

		for _, n := range g.Nodes {
			r := []string{n.ID}

			for _, e := range g.Edges {
				v := "0"

				if e.FromNodeID == n.ID {
					v = "-1"

					if !e.Directed {
						v = "1"
					}
				}

				if e.ToNodeID == n.ID {
					v = "1"
				}

				r = append(r, v)
			}

			if err := c.Write(r); err != nil {
				return err
			}
		}
	}

	c.Flush()

	return c.Error()
}

func safeCell(v string) string {

	if strings.HasPrefix(strings.TrimSpace(v), "=") || strings.HasPrefix(strings.TrimSpace(v), "+") || strings.HasPrefix(strings.TrimSpace(v), "-") || strings.HasPrefix(strings.TrimSpace(v), "@") {
		return "'" + v
	}

	return v
}

func geoJSON(w io.Writer, g logistics.LogisticsGraph) error {

	features := []any{}

	for _, n := range g.Nodes {
		features = append(features, map[string]any{"type": "Feature", "id": n.ID, "geometry": map[string]any{"type": "Point", "coordinates": []float64{n.Coordinates.Longitude, n.Coordinates.Latitude}}, "properties": n})
	}

	for _, e := range g.Edges {
		coords := [][2]float64{}

		for _, p := range e.Geometry {
			coords = append(coords, [2]float64{p.Longitude, p.Latitude})
		}

		features = append(features, map[string]any{"type": "Feature", "id": e.ID, "geometry": map[string]any{"type": "LineString", "coordinates": coords}, "properties": map[string]any{"from_node_id": e.FromNodeID, "to_node_id": e.ToNodeID, "mode": e.Mode, "distance_km": e.DistanceKM, "directed": e.Directed, "source": e.Source}})
	}

	return json.NewEncoder(w).Encode(map[string]any{"type": "FeatureCollection", "features": features})
}

type xmlData struct {
	Key   string `xml:"key,attr"`
	Value string `xml:",chardata"`
}

type xmlNode struct {
	ID   string    `xml:"id,attr"`
	Data []xmlData `xml:"data"`
}

type xmlEdge struct {
	ID       string    `xml:"id,attr"`
	From     string    `xml:"source,attr"`
	To       string    `xml:"target,attr"`
	Directed bool      `xml:"directed,attr"`
	Data     []xmlData `xml:"data"`
}

type xmlKey struct {
	ID   string `xml:"id,attr"`
	For  string `xml:"for,attr"`
	Name string `xml:"attr.name,attr"`
	Type string `xml:"attr.type,attr"`
}

func graphML(w io.Writer, g logistics.LogisticsGraph) error {

	document := struct {
		XMLName   xml.Name `xml:"graphml"`
		Namespace string   `xml:"xmlns,attr"`
		Keys      []xmlKey `xml:"key"`
		Graph     struct {
			ID      string    `xml:"id,attr"`
			Default string    `xml:"edgedefault,attr"`
			Nodes   []xmlNode `xml:"node"`
			Edges   []xmlEdge `xml:"edge"`
		} `xml:"graph"`
	}{Namespace: "http://graphml.graphdrawing.org/xmlns"}

	for _, name := range []string{"name", "longitude", "latitude", "modes", "external_ids", "mode", "distance_km", "geometry", "source"} {
		kind := "string"

		if name == "distance_km" || name == "longitude" || name == "latitude" {
			kind = "double"
		}

		document.Keys = append(document.Keys, xmlKey{name, "all", name, kind})
	}

	document.Graph.ID = g.ID
	document.Graph.Default = "directed"

	for _, n := range g.Nodes {
		modes, _ := json.Marshal(n.Modes)
		ids, _ := json.Marshal(n.ExternalIDs)
		document.Graph.Nodes = append(document.Graph.Nodes, xmlNode{n.ID, []xmlData{{"name", n.Name}, {"longitude", fmt.Sprint(n.Coordinates.Longitude)}, {"latitude", fmt.Sprint(n.Coordinates.Latitude)}, {"modes", string(modes)}, {"external_ids", string(ids)}}})
	}

	for _, e := range g.Edges {
		geometry, _ := json.Marshal(e.Geometry)
		document.Graph.Edges = append(document.Graph.Edges, xmlEdge{e.ID, e.FromNodeID, e.ToNodeID, e.Directed, []xmlData{{"mode", string(e.Mode)}, {"distance_km", fmt.Sprint(e.DistanceKM)}, {"geometry", string(geometry)}, {"source", e.Source}}})
	}

	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}

	return xml.NewEncoder(w).Encode(document)
}
